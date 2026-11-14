//go:build !test

package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/chanify/chanify/agentcli"
	"github.com/chanify/chanify/crypto"
	"github.com/chanify/chanify/pb"
	"google.golang.org/protobuf/proto"
)

func TestAgentProcessHelper(t *testing.T) {
	if os.Getenv("CHANIFY_AGENT_TEST_HELPER") != "1" {
		return
	}
	http.DefaultTransport = mcpTestTransport(func(r *http.Request) (*http.Response, error) {
		mode := os.Getenv("CHANIFY_AGENT_TEST_RESPONSE")
		if mode == "block" {
			if path := os.Getenv("CHANIFY_AGENT_TEST_READY"); path != "" {
				_ = os.WriteFile(path, []byte("ready"), 0600)
			}
			<-r.Context().Done()
			return nil, r.Context().Err()
		}
		body := `{"request-uid":"agent-process-request"}`
		status := 200
		switch mode {
		case "401":
			status = 401
			body = "private-server-secret"
		case "500":
			status = 500
			body = "private-server-secret"
		case "missing-id":
			body = `{}`
		}
		if r.URL.Path != "/v1/sender" {
			status = 500
		}
		if expected := os.Getenv("CHANIFY_AGENT_TEST_EXPECT_TEXT"); expected != "" {
			if err := r.ParseMultipartForm(8 << 20); err != nil || r.FormValue("text") != expected {
				status = 500
			}
		}
		if mode == "long-timeout" {
			deadline, ok := r.Context().Deadline()
			if !ok || time.Until(deadline) < 35*time.Second {
				status = 500
			}
		}
		return &http.Response{StatusCode: status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}, nil
	})
	var args []string
	if err := json.Unmarshal([]byte(os.Getenv("CHANIFY_AGENT_TEST_ARGS")), &args); err != nil {
		os.Exit(99)
	}
	rootCmd.SetArgs(args)
	Execute()
	os.Exit(0)
}
func agentConfig(t *testing.T) string {
	t.Helper()
	channel, _ := proto.Marshal(&pb.Channel{Type: pb.ChanType_User, Name: "dev"})
	payload, _ := proto.Marshal(&pb.Token{UserId: "synthetic-agent-user", Channel: channel, Expires: uint64(time.Now().Add(time.Hour).Unix())})
	raw := crypto.Base64Encode.EncodeToString(payload) + ".." + crypto.Base64Encode.EncodeToString(bytes.Repeat([]byte{3}, 32))
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte("client:\n  token: '"+raw+"'\n"), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}
func agentProcess(t *testing.T, ctx context.Context, args []string, mode string) *exec.Cmd {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(args)
	command := exec.CommandContext(ctx, executable, "-test.run=^TestAgentProcessHelper$")
	command.Env = append(os.Environ(), "CHANIFY_AGENT_TEST_HELPER=1", "CHANIFY_AGENT_TEST_ARGS="+string(encoded), "CHANIFY_AGENT_TEST_RESPONSE="+mode)
	return command
}
func runAgentProcess(t *testing.T, args []string, mode string, stdin io.Reader) (agentcli.Result, int, string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	command := agentProcess(t, ctx, args, mode)
	command.Stdin = stdin
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	exit := 0
	if err != nil {
		failure, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatal(err)
		}
		exit = failure.ExitCode()
	}
	var result agentcli.Result
	if json.Unmarshal(stdout.Bytes(), &result) != nil || result.SchemaVersion != 1 || strings.Count(stdout.String(), "\n") != 1 {
		t.Fatalf("invalid stdout: %q, stderr: %q", stdout.String(), stderr.String())
	}
	if strings.Contains(stdout.String()+stderr.String(), "private-server-secret") {
		t.Fatal("server error leaked")
	}
	return result, exit, stderr.String()
}
func TestAgentProcessJSONErrorsAndSuccess(t *testing.T) {
	config := agentConfig(t)
	base := []string{"--config", config, "agent", "send"}
	for _, tc := range []struct {
		args         []string
		mode         string
		exit         int
		status, code string
	}{
		{[]string{"--text", "@literal"}, "", 0, "submitted", ""},
		{[]string{"--text", "hello", "--timeout", "45s"}, "long-timeout", 0, "submitted", ""},
		{[]string{"--text", "hello", "--verbose"}, "401", 3, "failed", "unauthorized"},
		{[]string{"--text", "hello"}, "500", 5, "unknown", "outcome_unknown"},
		{[]string{"--text", "hello"}, "missing-id", 5, "unknown", "outcome_unknown"},
		{[]string{"--text", "hello", "--channel", "dev"}, "", 2, "failed", "channel_not_allowed"},
		{[]string{"--text", "hello", "--file", "missing"}, "", 2, "failed", "invalid_arguments"},
		{[]string{"--text=", "--stdin"}, "", 2, "failed", "invalid_arguments"},
		{[]string{"--text", ""}, "", 2, "failed", "invalid_arguments"},
		{[]string{"--stdin=false"}, "", 2, "failed", "invalid_arguments"},
		{[]string{"--text", "hello", "--timeout", "0s"}, "", 2, "failed", "invalid_arguments"},
		{[]string{"--text", "hello", "--timeout", "bad"}, "", 2, "failed", "invalid_arguments"},
		{[]string{"--text", "hello", "--token=hidden"}, "", 2, "failed", "invalid_arguments"},
		{[]string{"--text", "hello", "extra"}, "", 2, "failed", "invalid_arguments"},
		{nil, "", 2, "failed", "invalid_arguments"},
	} {
		args := append(append([]string{}, base...), tc.args...)
		result, exit, _ := runAgentProcess(t, args, tc.mode, nil)
		if exit != tc.exit || result.Status != tc.status || tc.code != "" && (result.Error == nil || result.Error.Code != tc.code) {
			t.Fatal(tc.args, result, exit)
		}
	}
}
func TestAgentProcessStrictConfiguration(t *testing.T) {
	invalid := filepath.Join(t.TempDir(), "invalid.yml")
	if err := os.WriteFile(invalid, []byte("client: [secret-config-data"), 0600); err != nil {
		t.Fatal(err)
	}
	empty := filepath.Join(t.TempDir(), "empty.yml")
	if err := os.WriteFile(empty, []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{invalid, empty, filepath.Join(t.TempDir(), "missing.yml")} {
		result, exit, stderr := runAgentProcess(t, []string{"--config", path, "agent", "send", "--text", "hello"}, "", nil)
		if exit != 2 || result.Error.Code != "config_error" || strings.Contains(stderr, "secret-config-data") {
			t.Fatal(result, exit, stderr)
		}
	}
}
func TestAgentProcessStdinAndTimeout(t *testing.T) {
	config := agentConfig(t)
	result, exit, _ := runAgentProcess(t, []string{"--config", config, "agent", "send", "--stdin"}, "", strings.NewReader("input from pipe\n"))
	if exit != 0 || result.Status != "submitted" {
		t.Fatal(result, exit)
	}
	result, exit, _ = runAgentProcess(t, []string{"--config", config, "agent", "send", "--text", "hello", "--timeout", "30ms"}, "block", nil)
	if exit != 5 || result.Status != "unknown" {
		t.Fatal(result, exit)
	}
}
func testAgentProcessSignalAfterSend(t *testing.T, signal os.Signal, expectedExit int) {
	config := agentConfig(t)
	ready := filepath.Join(t.TempDir(), "ready")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	command := agentProcess(t, ctx, []string{"--config", config, "agent", "send", "--text", "hello"}, "block")
	command.Env = append(command.Env, "CHANIFY_AGENT_TEST_READY="+ready)
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(ready); err == nil {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal("process did not send", stderr.String())
		case <-time.After(5 * time.Millisecond):
		}
	}
	if err := command.Process.Signal(signal); err != nil {
		t.Fatal(err)
	}
	err := command.Wait()
	failure, ok := err.(*exec.ExitError)
	if !ok || failure.ExitCode() != expectedExit {
		t.Fatal("wrong interrupt exit", err, stderr.String())
	}
	var result agentcli.Result
	if json.Unmarshal(stdout.Bytes(), &result) != nil || result.Status != "unknown" || result.Error.Code != "canceled" {
		t.Fatal(stdout.String(), stderr.String())
	}
}
func TestAgentHelpAndLegacySendCompatibility(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	help := agentProcess(t, ctx, []string{"agent", "send", "--help"}, "")
	output, err := help.Output()
	if err != nil || !strings.Contains(string(output), "--text-file") || strings.HasPrefix(string(output), "{") {
		t.Fatal("help unavailable", err)
	}
	config := agentConfig(t)
	path := filepath.Join(t.TempDir(), "legacy.txt")
	if err := os.WriteFile(path, []byte("legacy file text"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, input := range []string{"literal text", "@" + path} {
		command := agentProcess(t, ctx, []string{"--config", config, "send", "--text", input, "--channel", "CAE"}, "")
		expected := "literal text"
		if strings.HasPrefix(input, "@") {
			expected = "legacy file text"
		}
		command.Env = append(command.Env, "CHANIFY_AGENT_TEST_EXPECT_TEXT="+expected)
		output, err := command.Output()
		if err != nil || string(output) != "request-uid: agent-process-request\n" {
			t.Fatal("legacy send changed", err, string(output))
		}
	}
}

func TestAgentProcessStdinTimeoutBeforeSend(t *testing.T) {
	config := agentConfig(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	command := agentProcess(t, ctx, []string{"--config", config, "agent", "send", "--stdin", "--timeout", "30ms"}, "")
	input, err := command.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = input.Close() }()
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err = command.Run()
	failure, ok := err.(*exec.ExitError)
	if !ok || failure.ExitCode() != 4 {
		t.Fatal("stdin timeout exit", err, stderr.String())
	}
	var result agentcli.Result
	if json.Unmarshal(stdout.Bytes(), &result) != nil || result.Status != "failed" || result.Error.Code != "timeout" {
		t.Fatal(stdout.String())
	}
}

func TestAgentProcessSignalsAfterSend(t *testing.T) {
	t.Run("SIGINT", func(t *testing.T) { testAgentProcessSignalAfterSend(t, os.Interrupt, 130) })
	t.Run("SIGTERM", func(t *testing.T) { testAgentProcessSignalAfterSend(t, syscall.SIGTERM, 143) })
}
