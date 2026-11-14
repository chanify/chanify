//go:build !test

package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/chanify/chanify/agentcli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type agentCommandError struct {
	result agentcli.Result
	exit   int
}

func (e *agentCommandError) Error() string { return "agent send failed" }

type agentOutputError struct{}

func (*agentOutputError) Error() string { return "cannot write agent result" }
func isAgentSend(command *cobra.Command) bool {
	return command != nil && command.Annotations["chanify.agent.send"] == "true"
}
func writeAgentError(writer io.Writer, err error) int {
	var output *agentOutputError
	if errors.As(err, &output) {
		return 1
	}
	result, exit := agentcli.Failed("invalid_arguments", "Invalid command arguments. See chanify agent send --help.")
	var failure *agentCommandError
	if errors.As(err, &failure) {
		result, exit = failure.result, failure.exit
	}
	if json.NewEncoder(writer).Encode(result) != nil {
		return 1
	}
	return exit
}
func init() {
	parent := &cobra.Command{Use: "agent", Short: "Commands for coding agent automation", SilenceErrors: true, SilenceUsage: true}
	command := &cobra.Command{Use: "send", Short: "Send one notification and print one JSON result", Args: cobra.NoArgs, SilenceErrors: true, SilenceUsage: true, Annotations: map[string]string{"chanify.agent.send": "true"}, RunE: runAgentSend, Long: `Send one notification using client.endpoint and client.token from configuration.
Choose exactly one of --text, --text-file, --stdin, --image, or --file.
Text is literal: @ and - have no special meaning. File paths are local regular files.
Text accepts up to 400 Unicode characters (100,000 with file storage); title 80,
summary 200. Long text is attached in full. Summary does not replace short text.
PNG/JPEG images and files accept at most 4 MiB; images at most 32 million pixels.
The default service supports only short text. Self-hosted capabilities are checked.
Only server-wide sending tokens accept readable channel codes; omitted channel
uses sys.none. Ordinary sending tokens keep their signed destination.
Stdout is one schema_version=1 JSON result: submitted, failed, or unknown.
Exit codes: 0 submitted, 1 internal, 2 input/config, 3 authorization, 4 capability/
pre-send timeout, 5 unknown submission outcome, 130/143 interrupted. No auto retry.
Submitted means accepted, not delivered or read. --help prints this text instead.`}
	flags := command.Flags()
	flags.String("text", "", "Literal UTF-8 notification text.")
	flags.String("text-file", "", "Read UTF-8 text from a local file (prefer absolute paths).")
	flags.Bool("stdin", false, "Read UTF-8 text from redirected stdin until EOF.")
	flags.String("image", "", "PNG/JPEG image file path.")
	flags.String("file", "", "Attachment file path; basename becomes the filename.")
	flags.String("title", "", "Optional title, up to 80 Unicode characters.")
	flags.String("summary", "", "Long-text preview or attachment description, up to 200 characters.")
	flags.String("channel", "", "Readable channel code; only server-wide sending tokens accept this option.")
	flags.String("endpoint", "", "Override client.endpoint; default https://api.chanify.net.")
	flags.Duration("timeout", 30*time.Second, "Total input/discovery/send budget (positive duration).")
	parent.AddCommand(command)
	rootCmd.AddCommand(parent)
}
func runAgentSend(command *cobra.Command, _ []string) error {
	fail := func(code, message string) error {
		result, exit := agentcli.Failed(code, message)
		return &agentCommandError{result, exit}
	}
	flags := command.Flags()
	mode := ""
	count := 0
	for _, name := range []string{"text", "text-file", "stdin", "image", "file"} {
		if flags.Changed(name) {
			mode = name
			count++
		}
	}
	if count != 1 {
		return fail("invalid_arguments", "Choose exactly one of --text, --text-file, --stdin, --image, or --file.")
	}
	if mode == "stdin" {
		enabled, _ := flags.GetBool("stdin")
		if !enabled {
			return fail("invalid_arguments", "--stdin must be enabled when selected.")
		}
	}
	timeout, _ := flags.GetDuration("timeout")
	if timeout <= 0 {
		return fail("invalid_arguments", "--timeout must be a positive duration.")
	}
	if configLoadErr != nil {
		return fail("config_error", "Cannot read the configuration file.")
	}
	started := executionStarted
	if started.IsZero() {
		started = time.Now()
	}
	deadline := started.Add(timeout)
	ctx, cancel := context.WithDeadline(command.Context(), deadline)
	defer cancel()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)
	stop := make(chan struct{})
	defer close(stop)
	var interrupted atomic.Int32
	go func() {
		select {
		case received := <-signals:
			if received == syscall.SIGTERM {
				interrupted.Store(143)
			} else {
				interrupted.Store(130)
			}
			cancel()
		case <-stop:
		}
	}()
	source := ""
	if mode != "stdin" {
		source, _ = flags.GetString(mode)
	}
	title, _ := flags.GetString("title")
	summary, _ := flags.GetString("summary")
	var channel *string
	if flags.Changed("channel") {
		value, _ := flags.GetString("channel")
		channel = &value
	}
	endpoint := viper.GetString("client.endpoint")
	if flags.Changed("endpoint") {
		endpoint, _ = flags.GetString("endpoint")
		if endpoint == "" {
			return fail("config_error", "--endpoint must be a nonempty HTTP(S) URL.")
		}
	}
	opts := agentcli.Options{Mode: mode, Source: source, Title: title, Summary: summary, Channel: channel, Endpoint: endpoint, Token: viper.GetString("client.token"), Version: Version, Timeout: timeout}
	result, exit := agentcli.Run(ctx, opts, command.InOrStdin(), nil)
	if caught := interrupted.Load(); caught != 0 {
		result, exit = agentcli.Canceled(result, int(caught))
	}
	if exit != 0 {
		return &agentCommandError{result, exit}
	}
	if json.NewEncoder(command.OutOrStdout()).Encode(result) != nil {
		return &agentOutputError{}
	}
	return nil
}
