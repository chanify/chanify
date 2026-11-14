package core

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"mime/multipart"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chanify/chanify/crypto"
	"github.com/chanify/chanify/logic"
	"github.com/chanify/chanify/pb"
	"google.golang.org/protobuf/proto"
)

func TestServerTokenSender(t *testing.T) {
	c := New()
	defer c.Close()
	storage := t.TempDir()
	if err := c.Init(&logic.Options{DBUrl: "sqlite://?mode=memory", FilePath: storage, Registerable: true}); err != nil {
		t.Fatal(err)
	}
	key := crypto.GenerateSecretKey([]byte("synthetic-global-sender"))
	uid := key.ToID(0)
	user, err := c.logic.UpsertUser(uid, crypto.Base64Encode.EncodeToString(key.MarshalPublicKey()), false)
	if err != nil {
		t.Fatal(err)
	}
	makeToken := func(node string, expiry uint64) string {
		data, _ := proto.Marshal(&pb.Token{UserId: uid, NodeId: node, Expires: expiry, Scope: pb.TokenScope_ServerSend, TokenId: make([]byte, 16)})
		mac := hmac.New(sha256.New, user.SecretKey[:32])
		_, _ = mac.Write(data)
		return "cst1:" + crypto.Base64Encode.EncodeToString(data) + ".." + crypto.Base64Encode.EncodeToString(mac.Sum(nil))
	}
	token := makeToken(c.logic.NodeID, uint64(time.Now().Add(time.Hour).Unix()))
	channelData, _ := proto.Marshal(&pb.Channel{Type: pb.ChanType_User, Name: "development"})
	channel := crypto.Base64Encode.EncodeToString(channelData)
	if _, err := c.logic.ManageServerTokens(uid, "create", "", "Test sender", token); err != nil {
		t.Fatal(err)
	}
	handler := c.APIHandler()
	for _, tc := range []struct {
		name, method, path, ctype, body, token string
		status                                 int
	}{
		{"missing", "POST", "/v1/sender", "application/json", `{"text":"hello"}`, token, 404},

		{"empty query", "POST", "/v1/sender?channel=", "application/json", `{"text":"hello","channel":"` + channel + `"}`, token, 400},
		{"null", "POST", "/v1/sender", "application/json", `{"text":"hello","channel":null}`, token, 400},
		{"form missing", "POST", "/v1/sender", "application/x-www-form-urlencoded", url.Values{"text": {"hello"}, "token": {token}}.Encode(), "", 404},
		{"empty", "POST", "/v1/sender", "application/json", `{"text":"hello","channel":""}`, token, 400},
		{"invalid", "POST", "/v1/sender", "application/json", `{"text":"hello","channel":"bad!"}`, token, 400},
		{"json", "POST", "/v1/sender", "application/json", `{"text":"hello","channel":"` + channel + `"}`, token, 404},
		{"form", "POST", "/v1/sender", "application/x-www-form-urlencoded", url.Values{"text": {"hello"}, "channel": {channel}, "token": {token}}.Encode(), "", 404},
		{"plain", "POST", "/v1/sender?channel=" + channel, "text/plain", "hello", token, 404},
		{"get missing", "GET", "/v1/sender/" + token + "/hello", "", "", "", 404},
		{"get", "GET", "/v1/sender/" + token + "/hello?channel=" + channel, "", "", "", 404},
		{"node mismatch", "POST", "/v1/sender?channel=" + channel, "text/plain", "hello", makeToken("MFRGG", uint64(time.Now().Add(time.Hour).Unix())), 401},
		{"expired", "POST", "/v1/sender?channel=" + channel, "text/plain", "hello", makeToken(c.logic.NodeID, 1), 401},
		{"raw image missing", "POST", "/v1/sender", "image/png", "image", token, 404},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			req.Header.Set("Content-Type", tc.ctype)
			req.Header.Set("Token", tc.token)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			if w.Code != tc.status {
				t.Fatalf("got %d: %s", w.Code, w.Body.String())
			}
			// 404 here means validation passed and sending reached the local device lookup.
			if w.Code == 404 && !strings.Contains(w.Body.String(), "no devices found") {
				t.Fatal(w.Body.String())
			}
		})
	}
	for _, valid := range []bool{false, true} {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		_ = writer.WriteField("token", token)
		if valid {
			_ = writer.WriteField("channel", channel)
		} else {
			_ = writer.WriteField("channel", "")
		}
		part, _ := writer.CreateFormFile("file", "report.txt")
		_, _ = part.Write([]byte("report"))
		_ = writer.Close()
		req := httptest.NewRequest("POST", "/v1/sender", &body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		expected := 400
		if valid {
			expected = 404
		}
		if w.Code != expected {
			t.Fatalf("multipart: %d %s", w.Code, w.Body.String())
		}
		if !valid {
			files, _ := os.ReadDir(filepath.Join(storage, "files"))
			if len(files) != 0 {
				t.Fatal("invalid request saved attachment")
			}
		}
	}
	deviceKey := crypto.GenerateSecretKey([]byte("synthetic-global-device"))
	deviceHash := sha1.Sum(deviceKey.MarshalPublicKey())
	deviceID := strings.ToUpper(hex.EncodeToString(deviceHash[:]))
	if err := c.logic.BindDevice(uid, deviceID, deviceKey.EncodePublicKey(), 0); err != nil {
		t.Fatal(err)
	}
	if err := c.logic.UpdatePushToken(uid, deviceID, crypto.Base64Encode.EncodeToString([]byte("synthetic-push-token")), false); err != nil {
		t.Fatal(err)
	}
	previousPusher := logic.MockPusher
	logic.MockPusher = &MockAPNSPusher{}
	defer func() { logic.MockPusher = previousPusher }()
	req := httptest.NewRequest("POST", "/v1/sender", strings.NewReader(`{"channel":"`+channel+`","text":"completed"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Token", token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != 200 || !strings.Contains(w.Body.String(), "request-uid") {
		t.Fatalf("mock push failed: %d %s", w.Code, w.Body.String())
	}
	// A stateful node must still reject a user bound only for forwarding.
	if _, err := c.logic.UpsertUser(uid, key.EncodePublicKey(), true); err != nil {
		t.Fatal(err)
	}
	if _, err := c.parseToken(token); err == nil {
		t.Fatal("accepted forwarding user")
	}
	info, _ := c.logic.GetInfo()
	var node struct {
		Features []string `json:"features"`
	}
	_ = json.Unmarshal(info, &node)
	if !strings.Contains(strings.Join(node.Features, ","), "auth.server-token") {
		t.Fatal("missing feature")
	}
}

func TestServerTokenFeatureDisabledForForwarding(t *testing.T) {
	for _, opts := range []*logic.Options{{Secret: "synthetic-secret"}, {DBUrl: "nosql://?secret=synthetic-secret"}} {
		c := New()
		if err := c.Init(opts); err != nil {
			t.Fatal(err)
		}
		info, _ := c.logic.GetInfo()
		if strings.Contains(string(info), "auth.server-token") {
			t.Fatal("forwarding node advertised server token")
		}
		c.Close()
	}
}
