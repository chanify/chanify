//go:build !test

package cmd

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chanify/chanify/crypto"
	"github.com/chanify/chanify/pb"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/proto"
)

type mcpTestTransport func(*http.Request) (*http.Response, error)

func (f mcpTestTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestMCPStdioProcessHelper(t *testing.T) {
	if os.Getenv("CHANIFY_STDIO_TEST_HELPER") != "1" {
		return
	}
	http.DefaultTransport = mcpTestTransport(func(r *http.Request) (*http.Response, error) {
		status := 200
		body := `{"request-uid":"subprocess-request"}`
		if r.URL.Path != "/v1/sender" || r.Header.Get("token") == "" {
			status = 500
			body = `{}`
		}
		return &http.Response{StatusCode: status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}, nil
	})
	rootCmd.SetArgs([]string{"mcp", "--verbose", "--config", os.Getenv("CHANIFY_STDIO_TEST_CONFIG")})
	Execute()
	os.Exit(0)
}
func TestMCPStdioSubprocess(t *testing.T) {
	channel, _ := proto.Marshal(&pb.Channel{Type: pb.ChanType_Sys})
	payload, _ := proto.Marshal(&pb.Token{UserId: "synthetic-stdio-user", Channel: channel, Expires: uint64(time.Now().Add(time.Hour).Unix())})
	token := crypto.Base64Encode.EncodeToString(payload) + ".." + crypto.Base64Encode.EncodeToString(bytes.Repeat([]byte{7}, 32))
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte("client:\n  token: '"+token+"'\n"), 0600); err != nil {
		t.Fatal(err)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, executable, "-test.run=^TestMCPStdioProcessHelper$")
	command.Env = append(os.Environ(), "CHANIFY_STDIO_TEST_HELPER=1", "CHANIFY_STDIO_TEST_CONFIG="+path)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	session, err := mcp.NewClient(&mcp.Implementation{Name: "synthetic-agent", Version: "test"}, nil).Connect(ctx, &mcp.CommandTransport{Command: command}, &mcp.ClientSessionOptions{ProtocolVersion: "2025-11-25"})
	if err != nil {
		t.Fatal("stdio negotiation failed", err, stderr.String())
	}
	tools, err := session.ListTools(ctx, nil)
	if err != nil || len(tools.Tools) != 1 {
		t.Fatal("stdio discovery failed", err)
	}
	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "send_text", Arguments: map[string]string{"text": "hello"}})
	if err != nil || result.IsError {
		t.Fatal("stdio call failed", err)
	}
	if err := session.Close(); err != nil {
		t.Fatal("EOF did not stop subprocess", err)
	}
	if !strings.Contains(stderr.String(), "Using config file:") {
		t.Fatal("verbose configuration output did not go to stderr")
	}
	if strings.Contains(stderr.String(), token) {
		t.Fatal("credential leaked to stderr")
	}
}

func TestMCPBadConfigDoesNotLeakToStdout(t *testing.T) {
	path := filepath.Join(t.TempDir(), "invalid.yml")
	secret := "synthetic-config-secret"
	if err := os.WriteFile(path, []byte("client: ["+secret), 0600); err != nil {
		t.Fatal(err)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, executable, "-test.run=^TestMCPStdioProcessHelper$")
	command.Env = append(os.Environ(), "CHANIFY_STDIO_TEST_HELPER=1", "CHANIFY_STDIO_TEST_CONFIG="+path)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	stdout, err := command.Output()
	if err == nil {
		t.Fatal("invalid configuration did not stop MCP startup")
	}
	if len(stdout) != 0 || strings.Contains(stderr.String(), secret) {
		t.Fatal("configuration diagnostics polluted protocol or exposed secrets")
	}
}
