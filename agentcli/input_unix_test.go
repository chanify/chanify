//go:build unix

package agentcli

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func TestNamedPipeIsRejectedWithoutOpening(t *testing.T) {
	path := filepath.Join(t.TempDir(), "input.pipe")
	if err := unix.Mkfifo(path, 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	result, exit := Run(ctx, Options{Mode: "text-file", Source: path, Token: token(t, false)}, nil, nil)
	if exit != 2 || result.Error.Code != "file_read_failed" {
		t.Fatal(result, exit)
	}
}
