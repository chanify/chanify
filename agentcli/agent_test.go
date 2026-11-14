package agentcli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"image"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chanify/chanify/crypto"
	"github.com/chanify/chanify/mcpnotify"
	"github.com/chanify/chanify/pb"
	"google.golang.org/protobuf/proto"
)

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func token(t *testing.T, global bool) string {
	t.Helper()
	channel, _ := proto.Marshal(&pb.Channel{Type: pb.ChanType_User, Name: "dev"})
	payload := &pb.Token{UserId: "synthetic-user", NodeId: "synthetic-node", Channel: channel, Expires: uint64(time.Now().Add(time.Hour).Unix())}
	prefix := ""
	if global {
		prefix = "cst1:"
		payload.Scope = pb.TokenScope_ServerSend
		payload.TokenId = bytes.Repeat([]byte{1}, 16)
		payload.Channel = nil
	}
	data, err := proto.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	return prefix + crypto.Base64Encode.EncodeToString(data) + ".." + crypto.Base64Encode.EncodeToString(bytes.Repeat([]byte{2}, 32))
}
func response(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}
}
func TestTextInputsAndLiteralCharacters(t *testing.T) {
	text := "@literal\n报告\n"
	path := filepath.Join(t.TempDir(), "report.md")
	if err := os.WriteFile(path, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"text", "text-file", "stdin"} {
		t.Run(mode, func(t *testing.T) {
			source := text
			if mode == "text-file" {
				source = path
			}
			calls := 0
			client := &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
				calls++
				if r.Method != "POST" || r.URL.Path != "/v1/sender" {
					t.Error("unexpected request")
				}
				if err := r.ParseMultipartForm(8 << 20); err != nil {
					t.Fatal(err)
				}
				if r.FormValue("text") != text || r.FormValue("channel") != "" || r.Header.Get("token") == "" {
					t.Error("literal content or fixed channel changed")
				}
				return response(200, `{"request-uid":"synthetic-request"}`), nil
			})}
			result, exit := Run(context.Background(), Options{Mode: mode, Source: source, Token: token(t, false)}, strings.NewReader(text), client)
			if exit != 0 || result.Status != "submitted" || result.Format != "text" || result.RequestID != "synthetic-request" || calls != 1 {
				t.Fatal(result, exit)
			}
		})
	}
}
func TestLocalFailuresDoNotSend(t *testing.T) {
	directory := t.TempDir()
	empty := filepath.Join(directory, "empty")
	invalid := filepath.Join(directory, "invalid")
	huge := filepath.Join(directory, "huge")
	for path, data := range map[string][]byte{empty: {}, invalid: {0xff}, huge: bytes.Repeat([]byte{'x'}, 400001)} {
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	ch := "dev"
	missing := filepath.Join(directory, "missing")
	cases := []struct {
		opts  Options
		input io.Reader
		code  string
	}{
		{Options{Mode: "text", Source: "ok", Channel: &ch}, nil, "channel_not_allowed"},
		{Options{Mode: "text", Source: " "}, nil, "invalid_arguments"},
		{Options{Mode: "text", Source: string([]byte{0xff})}, nil, "invalid_utf8"},
		{Options{Mode: "text-file", Source: invalid}, nil, "invalid_utf8"},
		{Options{Mode: "text-file", Source: huge}, nil, "too_large"},
		{Options{Mode: "file", Source: empty}, nil, "invalid_arguments"},
		{Options{Mode: "file", Source: missing, Summary: "must not send"}, nil, "file_read_failed"},
		{Options{Mode: "image", Source: directory}, nil, "file_read_failed"},
		{Options{Mode: "stdin"}, strings.NewReader(string([]byte{0xff})), "invalid_utf8"},
		{Options{Mode: "stdin"}, strings.NewReader(strings.Repeat("a", 400001)), "too_large"},
		{Options{Mode: "text", Source: "ok", Title: strings.Repeat("a", 81)}, nil, "too_large"},
	}
	for _, entry := range cases {
		entry.opts.Token = token(t, false)
		result, exit := Run(context.Background(), entry.opts, entry.input, &http.Client{Transport: roundTrip(func(*http.Request) (*http.Response, error) {
			t.Error("local failure reached network")
			return nil, errors.New("unexpected")
		})})
		if exit != 2 || result.Status != "failed" || result.Error.Code != entry.code {
			t.Fatal(entry.code, result, exit)
		}
	}
}
func TestSubmissionOutcomeClassification(t *testing.T) {
	for _, tc := range []struct {
		status int
		body   string
		want   string
		exit   int
		code   string
	}{
		{200, `{"request-uid":"id"}`, "submitted", 0, ""},
		{401, `private-response`, "failed", 3, "unauthorized"},
		{403, `private-response`, "failed", 3, "unauthorized"},
		{413, `private-response`, "failed", 2, "too_large"},
		{400, `private-response`, "unknown", 5, "outcome_unknown"},
		{404, `private-response`, "unknown", 5, "outcome_unknown"},
		{429, `private-response`, "unknown", 5, "outcome_unknown"},
		{500, `private-response`, "unknown", 5, "outcome_unknown"},
		{200, `{}`, "unknown", 5, "outcome_unknown"},
		{200, `not-json`, "unknown", 5, "outcome_unknown"},
	} {
		count := 0
		result, exit := Run(context.Background(), Options{Mode: "text", Source: "hello", Token: token(t, false)}, nil, &http.Client{Transport: roundTrip(func(*http.Request) (*http.Response, error) { count++; return response(tc.status, tc.body), nil })})
		if result.Status != tc.want || exit != tc.exit || count != 1 {
			t.Fatal(tc, result, exit, count)
		}
		if result.Error != nil && result.Error.Code != tc.code {
			t.Fatal(tc, result)
		}
		encoded, _ := json.Marshal(result)
		if strings.Contains(string(encoded), "private-response") {
			t.Fatal("backend response leaked")
		}
	}
}
func TestTimeoutBeforeAndAfterSubmission(t *testing.T) {
	input, writer := io.Pipe()
	defer func() { _ = writer.Close() }()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	result, exit := Run(ctx, Options{Mode: "stdin", Token: token(t, false)}, input, nil)
	if exit != 4 || result.Status != "failed" || result.Error.Code != "timeout" {
		t.Fatal(result, exit)
	}
	ctx2, cancel2 := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel2()
	sends := 0
	result, exit = Run(ctx2, Options{Mode: "text", Source: "hello", Token: token(t, false)}, nil, &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		sends++
		<-r.Context().Done()
		return nil, r.Context().Err()
	})})
	if exit != 5 || result.Status != "unknown" || sends != 1 {
		t.Fatal(result, exit, sends)
	}
}
func TestDiscoveryAndAttachmentConversions(t *testing.T) {
	directory := t.TempDir()
	var pngData bytes.Buffer
	_ = png.Encode(&pngData, image.NewRGBA(image.Rect(0, 0, 1, 1)))
	long := strings.Repeat("报告", 1200)
	cases := []struct {
		mode, filename string
		data           []byte
		format         string
	}{
		{"text-file", "report.md", []byte(long), "text_attachment"}, {"image", "image.png", pngData.Bytes(), "image"}, {"file", "report.pdf", []byte("file bytes"), "file"},
	}
	for _, tc := range cases {
		path := filepath.Join(directory, tc.filename)
		if err := os.WriteFile(path, tc.data, 0600); err != nil {
			t.Fatal(err)
		}
		sends := 0
		channel := "development"
		result, exit := Run(context.Background(), Options{Mode: tc.mode, Source: path, Token: token(t, true), Endpoint: "https://node.example", Channel: &channel, Title: "Title", Summary: "Preview", Timeout: 45 * time.Second}, nil, &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
			if r.URL.Path == "/rest/v1/info" {
				return response(200, `{"nodeid":"synthetic-node","features":["auth.server-token","msg.file","msg.image"]}`), nil
			}
			sends++
			if err := r.ParseMultipartForm(8 << 20); err != nil {
				t.Fatal(err)
			}
			partName := "file"
			if tc.mode == "image" {
				partName = "image"
			}
			file, _, err := r.FormFile(partName)
			if err != nil {
				t.Fatal(err)
			}
			data, _ := io.ReadAll(file)
			_ = file.Close()
			if !bytes.Equal(data, tc.data) {
				t.Error("complete input bytes changed")
			}
			if r.FormValue("title") != "Title" || r.FormValue("text") != "Preview" || r.FormValue("channel") == "" {
				t.Error("missing caption/channel")
			}
			return response(200, `{"request-uid":"attachment-id"}`), nil
		})})
		if result.Status != "submitted" || result.Format != tc.format || exit != 0 || sends != 1 {
			t.Fatal(tc.mode, result, exit)
		}
	}
}
func TestDefaultLimitAndCompleteImageValidation(t *testing.T) {
	result, exit := Run(context.Background(), Options{Mode: "text", Source: strings.Repeat("汉", 401), Token: token(t, false)}, nil, nil)
	if result.Error.Code != "too_large" || exit != 2 {
		t.Fatal(result, exit)
	}
	var imageData bytes.Buffer
	_ = png.Encode(&imageData, image.NewRGBA(image.Rect(0, 0, 1, 1)))
	path := filepath.Join(t.TempDir(), "truncated.png")
	if err := os.WriteFile(path, imageData.Bytes()[:33], 0600); err != nil {
		t.Fatal(err)
	}
	result, exit = Run(context.Background(), Options{Mode: "image", Source: path, Token: token(t, false), Endpoint: "https://node.example"}, nil, &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		if r.Method != "GET" {
			t.Error("corrupt image submitted")
		}
		return response(200, `{"nodeid":"synthetic-node","features":["msg.image"]}`), nil
	})})
	if result.Error.Code != "invalid_image" || exit != 2 {
		t.Fatal(result, exit)
	}
}
func TestUnknownCancellationAndSuccessPriority(t *testing.T) {
	result, _ := FromError(&mcpnotify.SendError{Status: "unknown", Code: "outcome_unknown", Message: "unknown"})
	result, exit := Canceled(result, 143)
	if result.Status != "unknown" || result.Error.Code != "canceled" || exit != 143 {
		t.Fatal(result, exit)
	}
	result, exit = Canceled(Result{SchemaVersion: 1, Status: "submitted", RequestID: "id"}, 130)
	if result.Status != "submitted" || exit != 0 {
		t.Fatal(result, exit)
	}
}

func TestTotalBudgetIncludesInputAndDiscovery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	reader, writer := io.Pipe()
	go func() {
		defer func() { _ = writer.Close() }()
		timer := time.NewTimer(80 * time.Millisecond)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}
		_, _ = writer.Write([]byte("hello"))
	}()
	sends := 0
	result, exit := Run(ctx, Options{Mode: "stdin", Token: token(t, false), Endpoint: "https://node.example", Timeout: time.Second}, reader, &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		if r.Method != "GET" {
			sends++
			return response(200, `{"request-uid":"unexpected"}`), nil
		}
		timer := time.NewTimer(150 * time.Millisecond)
		defer timer.Stop()
		select {
		case <-r.Context().Done():
			return nil, r.Context().Err()
		case <-timer.C:
			return response(200, `{"nodeid":"synthetic-node"}`), nil
		}
	})})
	if exit != 4 || result.Status != "failed" || result.Error.Code != "timeout" || sends != 0 {
		t.Fatal("operation budget reset across stages", result, exit)
	}
}
func TestRegularSymlinkAndSizeBound(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "text")
	link := filepath.Join(directory, "link")
	if err := os.WriteFile(file, []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(file, link); err != nil {
		t.Skip("symlink unavailable", err)
	}
	data, err := readInput(context.Background(), Options{Mode: "text-file", Source: link}, nil)
	if err != nil || string(data) != "hello" {
		t.Fatal("regular symlink rejected", err)
	}
	if _, err := readBounded(context.Background(), strings.NewReader("grown input"), 3); err == nil {
		t.Fatal("bounded reader accepted growth past the limit")
	}
}
