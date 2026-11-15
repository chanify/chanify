//go:build !test

package cmd

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestLegacySendRequest(t *testing.T) {
	resetConfig(t)
	c := commandForTest(t, "send")
	dir := t.TempDir()
	textPath := filepath.Join(dir, "message.txt")
	attachment := filepath.Join(dir, "report.txt")
	for path, body := range map[string]string{textPath: "literal report", attachment: "complete attachment"} {
		if err := os.WriteFile(path, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != "POST" || r.URL.Path != "/v1/sender" {
			t.Error(r.Method, r.URL.Path)
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Error(err)
			return
		}
		defer func() { _ = r.MultipartForm.RemoveAll() }()
		for key, want := range map[string]string{"text": "literal report", "title": "Title", "token": "synthetic", "sound": "bell", "priority": "10", "interruption-level": "active", "channel": "dev", "copy": "copy me", "autocopy": "1"} {
			if got := r.FormValue(key); got != want {
				t.Errorf("%s=%q want %q", key, got, want)
			}
		}
		if got := r.MultipartForm.Value["action"]; len(got) != 2 || got[0] != "one" || got[1] != "two" {
			t.Error(got)
		}
		for _, key := range []string{"image", "audio", "file"} {
			f, _, err := r.FormFile(key)
			if err != nil {
				t.Error(err)
				continue
			}
			data, err := io.ReadAll(f)
			_ = f.Close()
			if err != nil || string(data) != "complete attachment" {
				t.Error(string(data), err)
			}
		}
		_, _ = io.WriteString(w, `{"request-uid":"receipt","res":200}`)
	}))
	defer srv.Close()
	for key, value := range map[string]interface{}{"client.token": "synthetic", "client.endpoint": srv.URL, "client.sound": "bell", "client.priority": 10, "client.interruption-level": "active", "client.autocopy": "1"} {
		viper.Set(key, value)
	}
	for flag, value := range map[string]string{"text": "@" + textPath, "title": "Title", "copy": "copy me", "channel": "dev", "image": attachment, "audio": attachment, "file": attachment} {
		setFlag(t, c, flag, value)
	}
	setFlag(t, c, "action", "one")
	setFlag(t, c, "action", "two")
	if err := runSendCmd(c, nil); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatal(calls)
	}
}

func TestLegacySendValidation(t *testing.T) {
	resetConfig(t)
	c := commandForTest(t, "send")
	if err := runSendCmd(c, nil); err == nil || err.Error() != "send token not found" {
		t.Fatal(err)
	}
	viper.Set("client.token", "synthetic")
	if err := runSendCmd(c, nil); err == nil || err.Error() != "no message content" {
		t.Fatal(err)
	}
	setFlag(t, c, "text", "hello")
	setFlag(t, c, "channel", "")
	if err := runSendCmd(c, nil); err == nil || err.Error() != "--channel cannot be empty" {
		t.Fatal(err)
	}
	setFlag(t, c, "channel", "channel")
	setFlag(t, c, "text", "@"+filepath.Join(t.TempDir(), "missing"))
	if err := runSendCmd(c, nil); err == nil {
		t.Fatal("missing text file accepted")
	}
	setFlag(t, c, "text", "hello")
	setFlag(t, c, "file", filepath.Join(t.TempDir(), "missing"))
	if err := runSendCmd(c, nil); err == nil {
		t.Fatal("missing attachment accepted")
	}
	var buffer bytes.Buffer
	if err := parseOtherMessage(&cobra.Command{}, multipart.NewWriter(&buffer)); err == nil {
		t.Fatal("missing file flag accepted")
	}
}

func TestTimelineFields(t *testing.T) {
	resetConfig(t)
	c := commandForTest(t, "send")
	timestamp := "2026-11-15T10:00:00.123+08:00"
	for _, tc := range []struct {
		args      []string
		timestamp string
		fail      bool
	}{
		{[]string{" load = 1.5 ", "zero="}, timestamp, false},
		{[]string{"missing-separator"}, timestamp, true},
		{[]string{" =value"}, timestamp, true},
		{nil, "invalid", true},
		{nil, "", false},
	} {
		setFlag(t, c, "timeline.timestamp", tc.timestamp)
		if err := c.Flags().Parse(tc.args); err != nil {
			t.Fatal(err)
		}
		var buffer bytes.Buffer
		writer := multipart.NewWriter(&buffer)
		err := parseTimelineMessage(c, "metrics", writer)
		if (err != nil) != tc.fail {
			t.Fatal(tc, err)
		}
		if tc.fail {
			continue
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		form, err := multipart.NewReader(&buffer, writer.Boundary()).ReadForm(1 << 20)
		if err != nil {
			t.Fatal(err)
		}
		if form.Value["timeline-code"][0] != "metrics" {
			t.Fatal(form.Value)
		}
		if len(tc.args) > 0 && (form.Value["timeline-items[load]"][0] != "1.5" || form.Value["timeline-items[zero]"][0] != "0") {
			t.Fatal(form.Value)
		}
	}
	setFlag(t, c, "timeline.code", "metrics")
	setFlag(t, c, "timeline.timestamp", "bad")
	viper.Set("client.token", "synthetic")
	if err := runSendCmd(c, nil); err == nil {
		t.Fatal("invalid timeline accepted")
	}
}

func TestSendMessageErrors(t *testing.T) {
	resetConfig(t)
	viper.Set("client.endpoint", "invalid")
	if err := sendMessage(context.Background(), &bytes.Buffer{}, "text/plain"); err == nil {
		t.Fatal("invalid endpoint accepted")
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(503) }))
	defer srv.Close()
	viper.Set("client.endpoint", srv.URL)
	if err := sendMessage(context.Background(), &bytes.Buffer{}, "text/plain"); err == nil {
		t.Fatal("node failure ignored")
	}
}

func TestInputFilesAndStdin(t *testing.T) {
	original := os.Stdin
	t.Cleanup(func() { os.Stdin = original })
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = r.Close() }()
	if _, err := w.Write([]byte("stdin text")); err != nil {
		t.Fatal(err)
	}
	_ = w.Close()
	os.Stdin = r
	data, name, err := readInputFile("-")
	if err != nil || string(data) != "stdin text" || name != "" {
		t.Fatal(string(data), name, err)
	}
	_ = r.Close()
	if _, _, err := readInputFile("-"); err == nil {
		t.Fatal("closed stdin accepted")
	}
	if data, name, err := readFile(""); err != nil || data != nil || name != "" {
		t.Fatal(data, name, err)
	}
	if got := fixFilename("", "fallback"); got != "fallback" {
		t.Fatal(got)
	}
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	setFieldValue(writer, "empty", nil)
	setFieldValues(writer, "empty", []string{""})
	setFieldValueInt(writer, "zero", 0)
	setFieldFile(writer, "empty", "", []byte("data"))
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(b.String(), "name=") {
		t.Fatal("empty field serialized")
	}
}
