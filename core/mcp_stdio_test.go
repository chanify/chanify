package core

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"image"
	"image/png"
	"io/fs"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chanify/chanify/agentcli"
	"github.com/chanify/chanify/crypto"
	"github.com/chanify/chanify/logic"
	"github.com/chanify/chanify/mcpnotify"
	"github.com/chanify/chanify/pb"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/proto"
)

type stdioNodeTransport struct{ handler http.Handler }

func (t stdioNodeTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	recorder := httptest.NewRecorder()
	t.handler.ServeHTTP(recorder, request)
	return recorder.Result(), nil
}
func TestStdioMCPWithNodeMCPDisabled(t *testing.T) {
	node := New()
	files := t.TempDir()
	if err := node.Init(&logic.Options{DBUrl: "sqlite://?mode=memory", Registerable: true, FilePath: files, Version: "1.5.0"}); err != nil {
		t.Fatal(err)
	}
	defer node.Close()
	handler := node.APIHandler()
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest("POST", "/mcp", nil))
	if w.Code != 404 {
		t.Fatal("test requires disabled node MCP")
	}
	key := crypto.GenerateSecretKey([]byte("synthetic-stdio-integration-user"))
	uid := key.ToID(0)
	user, err := node.logic.UpsertUser(uid, key.EncodePublicKey(), false)
	if err != nil {
		t.Fatal(err)
	}
	payload, _ := proto.Marshal(&pb.Token{Scope: pb.TokenScope_ServerSend, UserId: uid, NodeId: node.logic.NodeID, TokenId: bytes.Repeat([]byte{9}, 16), Expires: uint64(time.Now().Add(time.Hour).Unix())})
	signature := hmac.New(sha256.New, user.SecretKey[:32])
	_, _ = signature.Write(payload)
	token := "cst1:" + crypto.Base64Encode.EncodeToString(payload) + ".." + crypto.Base64Encode.EncodeToString(signature.Sum(nil))
	if _, err := node.logic.ManageServerTokens(uid, "create", "", "Local agent", token); err != nil {
		t.Fatal(err)
	}
	device := crypto.GenerateSecretKey([]byte("synthetic-stdio-device"))
	hash := sha1.Sum(device.MarshalPublicKey())
	deviceID := strings.ToUpper(hex.EncodeToString(hash[:]))
	if err := node.logic.BindDevice(uid, deviceID, device.EncodePublicKey(), 0); err != nil {
		t.Fatal(err)
	}
	if err := node.logic.UpdatePushToken(uid, deviceID, crypto.Base64Encode.EncodeToString([]byte("synthetic-push-token")), false); err != nil {
		t.Fatal(err)
	}
	previous := logic.MockPusher
	logic.MockPusher = &MockAPNSPusher{}
	defer func() { logic.MockPusher = previous }()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	local, err := mcpnotify.NewClient(ctx, "https://node.example", token, "test", &http.Client{Transport: stdioNodeTransport{handler}})
	if err != nil {
		t.Fatal(err)
	}
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	done := make(chan error, 1)
	go func() { done <- local.Server("test").Run(ctx, serverTransport) }()
	session, err := mcp.NewClient(&mcp.Implementation{Name: "synthetic-agent", Version: "test"}, nil).Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	tools, err := session.ListTools(ctx, nil)
	if err != nil || len(tools.Tools) != 3 {
		t.Fatal("node capability discovery failed", err)
	}
	var picture bytes.Buffer
	_ = png.Encode(&picture, image.NewRGBA(image.Rect(0, 0, 1, 1)))
	longText := strings.Repeat("报告正文", 900)
	for _, call := range []struct {
		name string
		args map[string]interface{}
	}{
		{"send_text", map[string]interface{}{"text": "hello"}},
		{"send_text", map[string]interface{}{"text": longText, "title": "Report", "summary": "report preview", "channel": "dev"}},
		{"send_image", map[string]interface{}{"data_base64": base64.StdEncoding.EncodeToString(picture.Bytes()), "title": "Image", "summary": "image caption"}},
		{"send_file", map[string]interface{}{"data_base64": base64.StdEncoding.EncodeToString([]byte("file bytes")), "filename": "report.txt", "title": "File", "summary": "file caption"}},
	} {
		result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: call.name, Arguments: call.args})
		if err != nil || result.IsError {
			encoded, _ := json.Marshal(result)
			t.Fatal(call.name, err, string(encoded))
		}
	}
	// The one-shot agent CLI uses the same backend even when node MCP is disabled.
	for _, entry := range []struct {
		mode, filename string
		data           []byte
		format         string
	}{
		{"text-file", "report.md", []byte(longText), "text_attachment"},
		{"image", "image.png", picture.Bytes(), "image"},
		{"file", "report.pdf", []byte("agent file bytes"), "file"},
	} {
		path := filepath.Join(t.TempDir(), entry.filename)
		if err := os.WriteFile(path, entry.data, 0600); err != nil {
			t.Fatal(err)
		}
		channel := "development"
		result, exit := agentcli.Run(ctx, agentcli.Options{Mode: entry.mode, Source: path, Channel: &channel, Title: "Agent report", Summary: "Agent preview", Token: token, Endpoint: "https://node.example", Timeout: time.Second}, nil, &http.Client{Transport: stdioNodeTransport{handler}})
		if exit != 0 || result.Status != "submitted" || result.Format != entry.format {
			t.Fatal("agent CLI integration", result, exit)
		}
	}
	found := false
	if err := filepath.WalkDir(files, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if bytes.Equal(data, []byte(longText)) {
				found = true
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if !found {
		t.Fatal("complete long text not preserved")
	}
	parsed, err := node.parseToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if err := node.logic.AdminUpdateToken("sending", uid, parsed.ServerTokenID(), "", true); err != nil {
		t.Fatal(err)
	}
	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "send_text", Arguments: map[string]string{"text": "revoked"}})
	if err != nil || !result.IsError {
		t.Fatal("revoked sending token accepted")
	}
	agentResult, agentExit := agentcli.Run(ctx, agentcli.Options{Mode: "text", Source: "revoked", Token: token, Endpoint: "https://node.example"}, nil, &http.Client{Transport: stdioNodeTransport{handler}})
	if agentExit != 3 || agentResult.Status != "failed" || agentResult.Error.Code != "unauthorized" {
		t.Fatal("agent accepted revoked token", agentResult, agentExit)
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
		t.Fatal("stdio session did not exit")
	}
}

func TestMultipartAttachmentWithoutStorageRejected(t *testing.T) {
	node := New()
	if err := node.Init(&logic.Options{DBUrl: "sqlite://?mode=memory", Registerable: true}); err != nil {
		t.Fatal(err)
	}
	defer node.Close()
	key := crypto.GenerateSecretKey([]byte("synthetic-no-storage-user"))
	uid := key.ToID(0)
	user, err := node.logic.UpsertUser(uid, key.EncodePublicKey(), false)
	if err != nil {
		t.Fatal(err)
	}
	payload, _ := proto.Marshal(&pb.Token{Scope: pb.TokenScope_ServerSend, UserId: uid, NodeId: node.logic.NodeID, TokenId: bytes.Repeat([]byte{4}, 16), Expires: uint64(time.Now().Add(time.Hour).Unix())})
	mac := hmac.New(sha256.New, user.SecretKey[:32])
	_, _ = mac.Write(payload)
	token := "cst1:" + crypto.Base64Encode.EncodeToString(payload) + ".." + crypto.Base64Encode.EncodeToString(mac.Sum(nil))
	if _, err := node.logic.ManageServerTokens(uid, "create", "", "Synthetic", token); err != nil {
		t.Fatal(err)
	}
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("text", "preview must not be submitted alone")
	part, err := writer.CreateFormFile("file", "report.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write([]byte("complete report"))
	_ = writer.Close()
	req := httptest.NewRequest("POST", "/v1/sender", &body)
	req.Header.Set("token", token)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	recorder := httptest.NewRecorder()
	node.APIHandler().ServeHTTP(recorder, req)
	if recorder.Code != 400 || !strings.Contains(recorder.Body.String(), "attachment storage is unavailable") {
		t.Fatal("missing attachment storage not rejected", recorder.Code)
	}
}
