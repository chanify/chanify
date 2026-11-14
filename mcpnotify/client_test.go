package mcpnotify

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"image"
	"image/png"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/chanify/chanify/crypto"
	"github.com/chanify/chanify/pb"
	"github.com/chanify/chanify/sender"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/proto"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func testToken(t *testing.T, global bool) string {
	t.Helper()
	channel, _ := ChannelData("dev")
	data := &pb.Token{UserId: "synthetic-user", NodeId: "synthetic-node", Expires: uint64(time.Now().Add(time.Hour).Unix()), Channel: channel}
	prefix := ""
	if global {
		data.Scope = pb.TokenScope_ServerSend
		data.TokenId = bytes.Repeat([]byte{1}, 16)
		data.Channel = nil
		prefix = "cst1:"
	}
	raw, err := proto.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	return prefix + crypto.Base64Encode.EncodeToString(raw) + ".." + crypto.Base64Encode.EncodeToString(bytes.Repeat([]byte{2}, 32))
}
func response(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}
}
func jsonArgs(value interface{}) json.RawMessage { data, _ := json.Marshal(value); return data }
func TestDefaultServerTextOnlyAndFixedChannel(t *testing.T) {
	raw := testToken(t, false)
	sends := 0
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.String() != sender.DefaultEndpoint+"/v1/sender" || r.Method != "POST" {
			t.Error("default node must not require discovery", r.URL)
		}
		if r.Header.Get("token") != raw || r.URL.RawQuery != "" {
			t.Error("credential not confined to request header")
		}
		if err := r.ParseMultipartForm(8 << 20); err != nil {
			t.Fatal(err)
		}
		if r.FormValue("text") != "hello" || r.FormValue("channel") != "" {
			t.Error("incorrect fixed-channel payload")
		}
		sends++
		return response(200, `{"request-uid":"synthetic-request"}`), nil
	})
	client, err := NewClient(context.Background(), "", raw, "test", &http.Client{Transport: transport})
	if err != nil {
		t.Fatal(err)
	}
	tools := Tools(client.caps)
	if len(tools) != 1 || strings.Contains(string(jsonArgs(tools[0].InputSchema)), `"channel"`) || !strings.Contains(tools[0].Description, "dev") {
		t.Fatal("fixed destination not described")
	}
	for _, args := range []map[string]interface{}{{"text": "hello", "channel": "sys.none"}, {"text": strings.Repeat("a", 401)}, {"text": "hello", "channel": nil}} {
		if result := client.Call(context.Background(), "send_text", jsonArgs(args)); !result.IsError {
			t.Fatal("invalid arguments accepted")
		}
	}
	if result := client.Call(context.Background(), "send_image", jsonArgs(map[string]string{"data_base64": "aGVsbG8="})); !result.IsError {
		t.Fatal("default node exposed images")
	}
	if result := client.Call(context.Background(), "send_text", jsonArgs(map[string]string{"text": "hello"})); result.IsError {
		t.Fatal(result)
	}
	if sends != 1 {
		t.Fatal("invalid calls reached network", sends)
	}
}
func TestCustomNodeCapabilitiesAndUploads(t *testing.T) {
	raw := testToken(t, true)
	text := strings.Repeat("汉", 1200)
	var picture bytes.Buffer
	_ = png.Encode(&picture, image.NewRGBA(image.Rect(0, 0, 1, 1)))
	expected := []struct {
		name, format, channel, file, text string
		data                              []byte
	}{
		{name: "send_text", format: "text", channel: "sys.none", text: "hello"},
		{name: "send_text", format: "text_attachment", channel: "dev", file: "file", text: "preview", data: []byte(text)},
		{name: "send_image", format: "image", channel: "dev", file: "image", text: "image summary", data: picture.Bytes()},
		{name: "send_file", format: "file", channel: "dev", file: "file", text: "file summary", data: []byte("file bytes")},
	}
	index := 0
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/prefix/rest/v1/info" {
			if r.Header.Get("token") != "" {
				t.Error("discovery leaked credential")
			}
			return response(200, `{"nodeid":"synthetic-node","features":["auth.server-token","msg.image","msg.file"]}`), nil
		}
		if index >= len(expected) {
			t.Fatal("unexpected send")
		}
		entry := expected[index]
		index++
		if r.URL.Path != "/prefix/v1/sender" || r.Header.Get("token") != raw {
			t.Fatal("wrong sending endpoint or credential")
		}
		if err := r.ParseMultipartForm(8 << 20); err != nil {
			t.Fatal(err)
		}
		channel, _ := ChannelData(entry.channel)
		if r.FormValue("channel") != crypto.Base64Encode.EncodeToString(channel) || r.FormValue("text") != entry.text || r.FormValue("title") != "title" {
			t.Error("incorrect fields")
		}
		if entry.file != "" {
			file, _, err := r.FormFile(entry.file)
			if err != nil {
				t.Fatal(err)
			}
			data, _ := io.ReadAll(file)
			_ = file.Close()
			if !bytes.Equal(data, entry.data) {
				t.Error("attachment bytes altered")
			}
		}
		return response(200, `{"request-uid":"upload-request"}`), nil
	})
	client, err := NewClient(context.Background(), "https://node.example/prefix/", raw, "test", &http.Client{Transport: transport})
	if err != nil {
		t.Fatal(err)
	}
	if len(Tools(client.caps)) != 3 {
		t.Fatal("capabilities missing")
	}
	for _, entry := range expected {
		args := map[string]interface{}{"title": "title"}
		if entry.channel != "sys.none" {
			args["channel"] = entry.channel
		}
		switch entry.name {
		case "send_text":
			if entry.file == "" {
				args["text"] = entry.text
			} else {
				args["text"] = text
				args["summary"] = entry.text
			}
		case "send_image", "send_file":
			args["data_base64"] = base64.StdEncoding.EncodeToString(entry.data)
			args["summary"] = entry.text
			if entry.name == "send_file" {
				args["filename"] = "report.txt"
			}
		}
		result := client.Call(context.Background(), entry.name, jsonArgs(args))
		if result.IsError || !strings.Contains(string(jsonArgs(result)), entry.format) {
			t.Fatal(result)
		}
	}
	if index != len(expected) {
		t.Fatal("not all uploads sent")
	}
}
func TestDiscoveryFailuresAndConservativeFallback(t *testing.T) {
	for _, global := range []bool{false, true} {
		for _, status := range []int{404, 401, 500} {
			client, err := NewClient(context.Background(), "https://node.example", testToken(t, global), "test", &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return response(status, `{"error":"private-server-detail"}`), nil
			})})
			if status == 404 && !global {
				if err != nil || len(Tools(client.caps)) != 1 {
					t.Fatal("legacy text fallback failed", err)
				}
			} else if err == nil {
				t.Fatal("failed discovery ignored")
			}
			if err != nil && strings.Contains(err.Error(), "private-server-detail") {
				t.Fatal("server response leaked")
			}
		}
	}
	for _, body := range []string{`{}`, `{"nodeid":"synthetic-node","features":[5]}`, `not-json`} {
		if _, err := NewClient(context.Background(), "https://node.example", testToken(t, false), "test", &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) { return response(200, body), nil })}); err == nil {
			t.Fatal("invalid node info accepted")
		}
	}
	if _, err := NewClient(context.Background(), "", testToken(t, true), "test", nil); err == nil {
		t.Fatal("global token accepted by default node")
	}
}
func TestClientSendCancellationAndRedactedErrors(t *testing.T) {
	raw := testToken(t, false)
	client, err := NewClient(context.Background(), "", raw, "test", &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) { <-r.Context().Done(); return nil, r.Context().Err() })})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result := client.Call(ctx, "send_text", jsonArgs(map[string]string{"text": "hello"}))
	if !result.IsError || !strings.Contains(string(jsonArgs(result)), "canceled") {
		t.Fatal("cancellation not propagated")
	}
	client.http, _ = sender.New("", "test", &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) { return response(401, raw), nil })})
	result = client.Call(context.Background(), "send_text", jsonArgs(map[string]string{"text": "hello"}))
	if !result.IsError || strings.Contains(string(jsonArgs(result)), raw) {
		t.Fatal("sending credential leaked")
	}
	client.http, _ = sender.New("", "test", &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) { return nil, errors.New(raw) })})
	result = client.Call(context.Background(), "send_text", jsonArgs(map[string]string{"text": "hello"}))
	if strings.Contains(string(jsonArgs(result)), raw) {
		t.Fatal("transport error leaked credential")
	}
}
func TestSDKStdioEquivalentConnection(t *testing.T) {
	client, err := NewClient(context.Background(), "", testToken(t, false), "test", &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return response(200, `{"request-uid":"sdk-request"}`), nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- client.Server("test").Run(ctx, serverTransport) }()
	session, err := mcp.NewClient(&mcp.Implementation{Name: "synthetic-agent", Version: "test"}, nil).Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	tools, err := session.ListTools(ctx, nil)
	if err != nil || len(tools.Tools) != 1 {
		t.Fatal("discovery failed", err)
	}
	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "send_text", Arguments: map[string]string{"text": "hello"}})
	if err != nil || result.IsError {
		t.Fatal("SDK send failed", err)
	}
	if err := session.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal("server did not exit on closed stream")
	}
}

func TestAttachmentCapabilitiesRecheckedBeforeSend(t *testing.T) {
	discoveries, sends := 0, 0
	client, err := NewClient(context.Background(), "https://node.example", testToken(t, false), "test", &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/rest/v1/info" {
			discoveries++
			if discoveries == 1 {
				return response(200, `{"nodeid":"synthetic-node","features":["msg.file"]}`), nil
			}
			return response(200, `{"nodeid":"synthetic-node","features":[]}`), nil
		}
		sends++
		return response(200, `{"request-uid":"unexpected"}`), nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	result := client.Call(context.Background(), "send_text", jsonArgs(map[string]string{"text": strings.Repeat("a", 2500)}))
	if !result.IsError || sends != 0 || discoveries != 2 {
		t.Fatal("long text could degrade to preview-only submission")
	}
}
