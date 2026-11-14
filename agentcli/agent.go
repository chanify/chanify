// Package agentcli adapts local inputs and submission outcomes for agent automation.
package agentcli

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/chanify/chanify/mcpnotify"
	"github.com/chanify/chanify/model"
	"github.com/mattn/go-isatty"
)

type Detail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
type Result struct {
	SchemaVersion int     `json:"schema_version"`
	Status        string  `json:"status"`
	Format        string  `json:"format,omitempty"`
	RequestID     string  `json:"request_id,omitempty"`
	Error         *Detail `json:"error,omitempty"`
}
type Options struct {
	Mode, Source, Title, Summary string
	Channel                      *string
	Endpoint, Token, Version     string
	Timeout                      time.Duration
}

func Failed(code, message string) (Result, int) {
	exit := 2
	switch code {
	case "invalid_token", "unauthorized":
		exit = 3
	case "unsupported", "discovery_failed", "timeout":
		exit = 4
	case "internal_error":
		exit = 1
	case "canceled":
		exit = 130
	}
	return Result{SchemaVersion: 1, Status: "failed", Error: &Detail{code, message}}, exit
}
func FromError(err error) (Result, int) {
	var failure *mcpnotify.SendError
	if errors.As(err, &failure) {
		if failure.Status == "unknown" {
			return Result{SchemaVersion: 1, Status: "unknown", Error: &Detail{"outcome_unknown", failure.Message}}, 5
		}
		return Failed(failure.Code, failure.Message)
	}
	return Failed("internal_error", "Cannot prepare notification.")
}
func Canceled(result Result, exit int) (Result, int) {
	if result.Status == "submitted" {
		return result, 0
	}
	message := "The operation was canceled before submission."
	if result.Status == "unknown" {
		message = "The operation was canceled. The notification may have been submitted; retrying may duplicate it."
	}
	result.Error = &Detail{"canceled", message}
	return result, exit
}
func inputError(code, message string) error {
	return &mcpnotify.SendError{Status: "failed", Code: code, Message: message}
}
func contextError(err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return inputError("timeout", "The operation timed out before submission.")
	}
	return inputError("canceled", "The operation was canceled before submission.")
}
func readBounded(ctx context.Context, reader io.Reader, limit int) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, contextError(err)
	}
	if reader == nil {
		return nil, inputError("file_read_failed", "Cannot read standard input.")
	}
	type readResult struct {
		data []byte
		err  error
	}
	ready := make(chan readResult, 1)
	go func() {
		data, err := io.ReadAll(io.LimitReader(reader, int64(limit)+1))
		ready <- readResult{data, err}
	}()
	select {
	case <-ctx.Done():
		if closer, ok := reader.(io.Closer); ok {
			_ = closer.Close()
		}
		return nil, contextError(ctx.Err())
	case read := <-ready:
		if err := ctx.Err(); err != nil {
			return nil, contextError(err)
		}
		if read.err != nil {
			return nil, inputError("file_read_failed", "Cannot read the input file.")
		}
		if len(read.data) > limit {
			return nil, inputError("too_large", "Input exceeds the size limit.")
		}
		return read.data, nil
	}
}
func readInput(ctx context.Context, opts Options, stdin io.Reader) ([]byte, error) {
	limit := mcpnotify.FileLimit
	if opts.Mode == "text" || opts.Mode == "text-file" || opts.Mode == "stdin" {
		limit = 400000
	}
	if opts.Mode == "text" {
		if len(opts.Source) > limit {
			return nil, inputError("too_large", "Text exceeds 400,000 UTF-8 bytes.")
		}
		return []byte(opts.Source), nil
	}
	if opts.Mode == "stdin" {
		if file, ok := stdin.(*os.File); ok && (isatty.IsTerminal(file.Fd()) || isatty.IsCygwinTerminal(file.Fd())) {
			return nil, inputError("invalid_arguments", "--stdin requires redirected input, not a terminal.")
		}
		return readBounded(ctx, stdin, limit)
	}
	info, err := os.Stat(opts.Source)
	if err != nil || !info.Mode().IsRegular() {
		return nil, inputError("file_read_failed", "Input must be a readable regular file.")
	}
	if info.Size() > int64(limit) {
		return nil, inputError("too_large", "Input file exceeds the size limit.")
	}
	file, err := os.Open(opts.Source)
	if err != nil {
		return nil, inputError("file_read_failed", "Cannot open the input file.")
	}
	defer func() { _ = file.Close() }()
	info, err = file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return nil, inputError("file_read_failed", "Input must be a readable regular file.")
	}
	return readBounded(ctx, file, limit)
}
func Run(ctx context.Context, opts Options, stdin io.Reader, transport *http.Client) (Result, int) {
	if err := ctx.Err(); err != nil {
		return FromError(contextError(err))
	}
	switch opts.Mode {
	case "text", "text-file", "stdin", "image", "file":
	default:
		return Failed("invalid_arguments", "Choose exactly one content input.")
	}
	if opts.Mode != "stdin" && opts.Source == "" {
		return Failed("invalid_arguments", "Content input cannot be empty.")
	}
	for _, text := range []string{opts.Source, opts.Title, opts.Summary} {
		if !utf8.ValidString(text) {
			return Failed("invalid_utf8", "Text arguments must contain valid UTF-8.")
		}
	}
	if opts.Token == "" {
		return Failed("config_error", "Configure client.token with a sending credential.")
	}
	token, err := model.ParseToken(opts.Token)
	if err != nil || token.IsExpires() {
		return Failed("invalid_token", "Configure a valid, unexpired sending token.")
	}
	if opts.Channel != nil {
		if !token.IsServerToken() {
			return Failed("channel_not_allowed", "The configured token has a fixed channel; omit --channel.")
		}
		if _, err := mcpnotify.ChannelData(*opts.Channel); err != nil {
			return Failed("invalid_channel", "Provide a valid channel code.")
		}
	}
	if utf8.RuneCountInString(opts.Title) > 80 || utf8.RuneCountInString(opts.Summary) > 200 {
		return Failed("too_large", "Title accepts 80 and summary 200 Unicode characters.")
	}
	data, err := readInput(ctx, opts, stdin)
	if err != nil {
		return FromError(err)
	}
	if len(data) == 0 {
		return Failed("invalid_arguments", "Content input cannot be empty.")
	}
	args := map[string]interface{}{"title": opts.Title, "summary": opts.Summary}
	if opts.Channel != nil {
		args["channel"] = *opts.Channel
	}
	name := "send_text"
	switch opts.Mode {
	case "text", "text-file", "stdin":
		if !utf8.Valid(data) {
			return Failed("invalid_utf8", "Text input must contain valid UTF-8.")
		}
		if strings.TrimSpace(string(data)) == "" {
			return Failed("invalid_arguments", "Text must contain non-whitespace content.")
		}
		args["text"] = string(data)
	case "image", "file":
		name = "send_" + opts.Mode
		args["data_base64"] = base64.StdEncoding.EncodeToString(data)
		if opts.Mode == "file" {
			args["filename"] = filepath.Base(opts.Source)
		}
	}
	client, err := mcpnotify.NewClientWithTimeout(ctx, opts.Endpoint, opts.Token, opts.Version, transport, opts.Timeout)
	if err != nil {
		return FromError(err)
	}
	raw, err := json.Marshal(args)
	if err != nil {
		return Failed("internal_error", "Cannot encode notification.")
	}
	submission, err := client.Send(ctx, name, raw)
	if err != nil {
		return FromError(err)
	}
	return Result{SchemaVersion: 1, Status: "submitted", Format: submission.Format, RequestID: submission.RequestID}, 0
}
