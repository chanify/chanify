package mcpnotify

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/chanify/chanify/crypto"
	"github.com/chanify/chanify/model"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/chanify/chanify/pb"
	"google.golang.org/protobuf/proto"
)

func TestChannelCodes(t *testing.T) {
	for _, tc := range []struct {
		code    string
		kind    pb.ChanType
		channel pb.ChanCode
	}{{"sys.none", pb.ChanType_Sys, pb.ChanCode_Uncategorized}, {"sys.timesets", pb.ChanType_Sys, pb.ChanCode_TimeSets}, {"team.build-1", pb.ChanType_User, 0}} {
		data, err := ChannelData(tc.code)
		if err != nil {
			t.Fatal(err)
		}
		var ch pb.Channel
		if err := proto.Unmarshal(data, &ch); err != nil || ch.Type != tc.kind || ch.Code != tc.channel {
			t.Fatal(&ch, err)
		}
	}
	for _, code := range []string{"", "sys.device", "sys.unknown", "a/b", strings.Repeat("x", 65)} {
		if _, err := ChannelData(code); err == nil {
			t.Fatalf("accepted %q", code)
		}
	}
}

func TestArgumentValidationFailures(t *testing.T) {
	caps := Capabilities{Images: true, Files: true, SelectableChannel: true}
	for _, tc := range []struct {
		name string
		raw  string
		code string
	}{
		{"send_text", `{"text":1}`, "invalid_arguments"},
		{"send_text", `{"text":"ok","title":null}`, "invalid_arguments"},
		{"send_text", `{"text":"ok","channel":"sys.device"}`, "invalid_channel"},
		{"send_text", `{"text":"ok","title":"` + strings.Repeat("x", 81) + `"}`, "too_large"},
		{"send_text", `{"text":"ok","summary":"` + strings.Repeat("x", 201) + `"}`, "too_large"},
		{"send_text", `{"text":"` + strings.Repeat("x", 100001) + `"}`, "too_large"},
		{"send_file", `{"data_base64":"` + strings.Repeat("A", base64.StdEncoding.EncodedLen(FileLimit)+1) + `","filename":"a"}`, "too_large"},
		{"send_file", `{"data_base64":"!","filename":"a"}`, "invalid_data"},
		{"send_file", `{"data_base64":"YQ==","filename":"../a"}`, "invalid_filename"},
		{"send_image", `{"data_base64":"YQ=="}`, "invalid_image"},
		{"unknown", `{}`, "unsupported"},
	} {
		t.Run(tc.code+"/"+tc.name, func(t *testing.T) {
			_, err := Parse(tc.name, json.RawMessage(tc.raw), caps)
			var validation *ValidationError
			if !errors.As(err, &validation) || validation.Code != tc.code || err.Error() == "" {
				t.Fatal(err)
			}
			if result := ValidationFailure(err); !result.IsError {
				t.Fatal(result)
			}
		})
	}
	if result := ValidationFailure(io.EOF); !result.IsError {
		t.Fatal(result)
	}
	if got := Preview(&Arguments{Text: "text"}); got != "text" {
		t.Fatal(got)
	}
	for _, caps := range []Capabilities{{Images: true}, {Files: true}} {
		if tools := Tools(caps); len(tools) != 2 {
			t.Fatal(tools)
		}
	}
}

func TestSendErrorPreservesCause(t *testing.T) {
	for _, err := range []error{preflightError(context.Canceled), preflightError(context.DeadlineExceeded), sendingError(io.EOF)} {
		var result *SendError
		if !errors.As(err, &result) || !errors.Is(err, result.Cause) || err.Error() == "" {
			t.Fatal(err)
		}
	}
	if !errors.Is(sendingError(io.EOF), io.EOF) {
		t.Fatal("cause lost")
	}
}

func TestFixedTokenChannelValidation(t *testing.T) {
	for _, tc := range []struct {
		channel []byte
		valid   bool
		name    string
	}{
		{[]byte{0xff}, false, ""},
		{[]byte{0x28, 1}, false, ""},
		{mustChannel(t, &pb.Channel{Type: pb.ChanType_User}), false, ""},
		{mustChannel(t, &pb.Channel{Type: pb.ChanType_None}), false, ""},
		{mustChannel(t, &pb.Channel{Type: pb.ChanType_Sys, Code: pb.ChanCode_Device}), false, ""},
		{mustChannel(t, &pb.Channel{Type: pb.ChanType_Sys, Code: pb.ChanCode_TimeSets}), true, "sys.timesets"},
		{mustChannel(t, &pb.Channel{Type: pb.ChanType_Sys, Code: pb.ChanCode_Uncategorized}), true, "sys.none"},
	} {
		payload := &pb.Token{UserId: "synthetic", Channel: tc.channel, Expires: uint64(time.Now().Add(time.Hour).Unix())}
		payload.Expires = uint64(time.Now().Add(time.Hour).Unix())
		data, err := proto.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		raw := crypto.Base64Encode.EncodeToString(data) + ".." + crypto.Base64Encode.EncodeToString(make([]byte, 32))
		client, err := NewClientWithTimeout(context.Background(), "", raw, "test", nil, time.Second)
		if (err == nil) != tc.valid {
			t.Fatal(tc, err)
		}
		if tc.valid && client.caps.FixedChannel != tc.name {
			t.Fatal(client.caps)
		}
	}
	for _, payload := range []*pb.Token{{UserId: "u", DeviceId: []byte("device")}, {UserId: "u", DataHash: []byte("hash")}, {}} {
		payload.Expires = uint64(time.Now().Add(time.Hour).Unix())
		data, err := proto.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		raw := crypto.Base64Encode.EncodeToString(data) + ".." + crypto.Base64Encode.EncodeToString(make([]byte, 32))
		if _, err := NewClient(context.Background(), "", raw, "test", nil); err == nil {
			t.Fatal("non-sending token accepted")
		}
	}
}
func mustChannel(t *testing.T, ch *pb.Channel) []byte {
	t.Helper()
	data, err := proto.Marshal(ch)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

type failWrite struct{ calls, failAt int }

func (w *failWrite) Write(p []byte) (int, error) {
	w.calls++
	if w.calls == w.failAt {
		return 0, io.ErrClosedPipe
	}
	return len(p), nil
}

func TestMultipartEncodingFailures(t *testing.T) {
	for _, tc := range []struct {
		name       string
		args       *Arguments
		selectable bool
		failAt     int
	}{
		{"send_text", &Arguments{Channel: "sys.none", Text: "text"}, true, 1},
		{"send_text", &Arguments{Title: "title", Text: "text"}, false, 1},
		{"send_file", &Arguments{Filename: "report", Bytes: []byte("data")}, false, 1},
		{"send_file", &Arguments{Filename: "report", Bytes: []byte("data")}, false, 2},
		{"send_text", &Arguments{Text: "text"}, false, 1},
		{"send_text", &Arguments{Text: strings.Repeat("x", 2000)}, false, 3},
		{"send_text", &Arguments{Text: "text"}, false, 3},
	} {
		writer := multipart.NewWriter(&failWrite{failAt: tc.failAt})
		if _, err := encodeNotification(writer, tc.name, tc.args, tc.selectable); err == nil {
			t.Fatal("encoding failure ignored", tc.failAt, tc.name)
		}
	}
}

func TestFailureAdaptersAndExpiredClient(t *testing.T) {
	if result := callFailure(io.EOF); !result.IsError {
		t.Fatal(result)
	}
	err := sendingValidationError(io.EOF)
	var failure *SendError
	if !errors.As(err, &failure) || failure.Code != "invalid_arguments" {
		t.Fatal(err)
	}
	c, err := NewClient(context.Background(), "", testToken(t, false), "test", nil)
	if err != nil {
		t.Fatal(err)
	}
	payload, err := proto.Marshal(&pb.Token{UserId: "synthetic", Expires: 1})
	if err != nil {
		t.Fatal(err)
	}
	raw := crypto.Base64Encode.EncodeToString(payload) + ".." + crypto.Base64Encode.EncodeToString(make([]byte, 32))
	c.parsed, err = model.ParseToken(raw)
	if err != nil {
		t.Fatal(err)
	}
	result := c.Call(context.Background(), "send_text", json.RawMessage(`{"text":"hello"}`))
	if !result.IsError || result.StructuredContent.(map[string]interface{})["code"] != "unauthorized" {
		t.Fatal(result)
	}
	if _, err := Parse("send_text", json.RawMessage(`[]`), Capabilities{}); err == nil {
		t.Fatal("array arguments accepted")
	}
	c.parsed, err = model.ParseToken(testToken(t, false))
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.Send(context.Background(), "send_text", json.RawMessage(`{"text":" "}`))
	if !errors.As(err, &failure) || failure.Code != "invalid_arguments" {
		t.Fatal(err)
	}
}

func TestAttachmentRecheckFailures(t *testing.T) {
	for _, mode := range []string{"discovery", "identity"} {
		calls := 0
		transport := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			calls++
			if calls == 1 {
				return response(200, `{"nodeid":"synthetic-node","features":["msg.file","auth.server-token"]}`), nil
			}
			if mode == "discovery" {
				return response(503, `{}`), nil
			}
			return response(200, `{"nodeid":"changed-node","features":["msg.file","auth.server-token"]}`), nil
		})}
		c, err := NewClient(context.Background(), "https://node.example", testToken(t, true), "test", transport)
		if err != nil {
			t.Fatal(err)
		}
		_, err = c.Send(context.Background(), "send_file", json.RawMessage(`{"data_base64":"YQ==","filename":"report.txt"}`))
		var failure *SendError
		if !errors.As(err, &failure) {
			t.Fatal(err)
		}
		want := "discovery_failed"
		if mode == "identity" {
			want = "unauthorized"
		}
		if failure.Code != want || calls != 2 {
			t.Fatal(failure, calls)
		}
	}
	if _, err := NewClient(context.Background(), "invalid", testToken(t, false), "test", nil); err == nil {
		t.Fatal("invalid endpoint accepted")
	}
}

func TestEncodingFailureDoesNotSubmit(t *testing.T) {
	c, err := NewClient(context.Background(), "", testToken(t, false), "test", nil)
	if err != nil {
		t.Fatal(err)
	}
	c.newWriter = func(io.Writer) *multipart.Writer { return multipart.NewWriter(&failWrite{failAt: 1}) }
	_, err = c.Send(context.Background(), "send_text", json.RawMessage(`{"text":"message"}`))
	var failure *SendError
	if !errors.As(err, &failure) || failure.Code != "internal_error" || failure.Status != "failed" {
		t.Fatal(err)
	}
}
