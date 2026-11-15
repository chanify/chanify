package agentcli

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFailureExitCodes(t *testing.T) {
	for _, tc := range []struct {
		code string
		exit int
	}{{"invalid_token", 3}, {"unauthorized", 3}, {"unsupported", 4}, {"discovery_failed", 4}, {"timeout", 4}, {"internal_error", 1}, {"canceled", 130}, {"invalid_arguments", 2}} {
		r, exit := Failed(tc.code, "message")
		if exit != tc.exit || r.SchemaVersion != 1 || r.Status != "failed" || r.Error.Code != tc.code || r.Error.Message != "message" {
			t.Fatal(r, exit)
		}
	}
	r, exit := FromError(io.EOF)
	if exit != 1 || r.Error.Code != "internal_error" {
		t.Fatal(r, exit)
	}
}

type readFailure struct{}

func (readFailure) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }
func TestInputReaderFailures(t *testing.T) {
	for _, reader := range []io.Reader{nil, readFailure{}} {
		_, err := readBounded(context.Background(), reader, 10)
		r, _ := FromError(err)
		if r.Error.Code != "file_read_failed" {
			t.Fatal(r)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := readBounded(ctx, strings.NewReader("ok"), 10); err == nil {
		t.Fatal("cancellation ignored")
	}
	ctx, cancel = context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()
	r, exit := Run(ctx, Options{}, nil, nil)
	if exit != 4 || r.Error.Code != "timeout" {
		t.Fatal(r, exit)
	}
	if _, err := readInput(context.Background(), Options{Mode: "text", Source: strings.Repeat("x", 400001)}, nil); err == nil {
		t.Fatal("oversized literal accepted")
	}
}
func TestRunPreflightFailures(t *testing.T) {
	badChannel := "sys.device"
	for _, tc := range []struct {
		opts Options
		code string
	}{
		{Options{}, "invalid_arguments"},
		{Options{Mode: "text"}, "invalid_arguments"},
		{Options{Mode: "text", Source: "text"}, "config_error"},
		{Options{Mode: "text", Source: "text", Token: "invalid"}, "invalid_token"},
		{Options{Mode: "text", Source: "text", Token: token(t, true), Channel: &badChannel}, "invalid_channel"},
		{Options{Mode: "text", Source: "text", Token: token(t, false), Endpoint: "invalid"}, "config_error"},
	} {
		r, _ := Run(context.Background(), tc.opts, nil, &http.Client{Transport: roundTrip(func(*http.Request) (*http.Response, error) {
			t.Error("preflight reached network")
			return nil, errors.New("unexpected")
		})})
		if r.Error == nil || r.Error.Code != tc.code {
			t.Fatalf("want %s: %+v", tc.code, r)
		}
	}
}

type invalidStatFile struct {
	inputFile
	info fs.FileInfo
	err  error
}

func (f invalidStatFile) Stat() (fs.FileInfo, error) { return f.info, f.err }

func TestInputOpenAndStatFailures(t *testing.T) {
	path := filepath.Join(t.TempDir(), "input.txt")
	if err := os.WriteFile(path, []byte("data"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"open", "stat", "directory"} {
		files := systemInputs
		files.Open = func(path string) (inputFile, error) {
			if mode == "open" {
				return nil, io.ErrClosedPipe
			}
			f, err := os.Open(path)
			if err != nil {
				return nil, err
			}
			if mode == "stat" {
				return invalidStatFile{inputFile: f, err: io.ErrClosedPipe}, nil
			}
			info, err := os.Stat(filepath.Dir(path))
			return invalidStatFile{inputFile: f, info: info, err: err}, nil
		}
		_, err := readInputWithFiles(context.Background(), Options{Mode: "text-file", Source: path}, nil, &files)
		if err == nil {
			t.Fatal("input failure ignored", mode)
		}
	}
	files := systemInputs
	files.Terminal = func(*os.File) bool { return true }
	if _, err := readInputWithFiles(context.Background(), Options{Mode: "stdin"}, os.Stdin, &files); err == nil {
		t.Fatal("terminal stdin accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := boundedResult(ctx, []byte("data"), nil, 10); err == nil {
		t.Fatal("cancellation after read ignored")
	}
}
