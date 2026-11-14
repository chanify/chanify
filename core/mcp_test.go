package core

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/chanify/chanify/crypto"
	"github.com/chanify/chanify/logic"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func mcpFixture(t *testing.T, files, listen bool) (*Core, string, string) {
	t.Helper()
	c := New()
	opts := &logic.Options{DBUrl: "sqlite://?mode=memory", Registerable: true, MCPEnabled: true, MCPListen: listen, Version: "1.5.0"}
	if files {
		opts.FilePath = t.TempDir()
	}
	if err := c.Init(opts); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(c.Close)
	key := crypto.GenerateSecretKey([]byte("synthetic-mcp-account"))
	uid := key.ToID(0)
	if _, err := c.logic.UpsertUser(uid, key.EncodePublicKey(), false); err != nil {
		t.Fatal(err)
	}
	record, raw, err := c.logic.CreateMCPToken(uid, "Agent", time.Now().Add(time.Hour).Unix())
	if err != nil {
		t.Fatal(err)
	}
	return c, raw, record.ID
}
func mcpRPC(t *testing.T, h http.Handler, raw, method string, params interface{}) (int, map[string]interface{}) {
	t.Helper()
	body, _ := json.Marshal(map[string]interface{}{"jsonrpc": "2.0", "id": 1, "method": method, "params": params})
	req := httptest.NewRequest("POST", "http://localhost/mcp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("MCP-Protocol-Version", "2025-11-25")
	req.Header.Set("Authorization", "Bearer "+raw)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	var result map[string]interface{}
	if w.Code == 200 {
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatalf("%s", w.Body.String())
		}
	}
	return w.Code, result
}
func TestMCPDiscoveryAndValidation(t *testing.T) {
	for _, files := range []bool{false, true} {
		c, raw, _ := mcpFixture(t, files, false)
		h := c.APIHandler()
		status, response := mcpRPC(t, h, raw, "tools/list", map[string]interface{}{})
		if status != 200 {
			t.Fatal(status, response)
		}
		tools := response["result"].(map[string]interface{})["tools"].([]interface{})
		expected := 1
		if files {
			expected = 3
		}
		if len(tools) != expected {
			t.Fatal(response)
		}
		for _, item := range tools {
			tool := item.(map[string]interface{})
			if tool["name"] != "list_channels" {
				schema := tool["inputSchema"].(map[string]interface{})
				if strings.Contains(string(mustJSON(schema["required"])), "channel") {
					t.Fatal("channel must be optional")
				}
			}
		}
		for _, tc := range []struct {
			name string
			args map[string]interface{}
		}{
			{"send_text", map[string]interface{}{"text": "hello", "channel": ""}},
			{"send_text", map[string]interface{}{"text": "hello", "channel": nil}},
			{"send_text", map[string]interface{}{"text": "hello", "channel": "bad!"}},
			{"send_text", map[string]interface{}{"text": strings.Repeat("x", 100001), "channel": "development"}},
			{"send_text", map[string]interface{}{"text": "hello", "channel": "development", "unexpected": true}},
		} {
			status, result := mcpRPC(t, h, raw, "tools/call", map[string]interface{}{"name": tc.name, "arguments": tc.args})
			if status != 200 || result["result"].(map[string]interface{})["isError"] != true {
				t.Fatal(result)
			}
		}

		req := httptest.NewRequest("POST", "http://localhost/mcp", strings.NewReader(`{}`))
		req.Header.Set("Origin", "https://evil.invalid")
		req.Header.Set("Authorization", "Bearer "+raw)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if w.Code != 403 {
			t.Fatal("cross origin accepted")
		}
		if status, _ := mcpRPC(t, h, "cst1:not-a-mcp-token", "tools/list", map[string]interface{}{}); status != 401 {
			t.Fatal("sender token accepted")
		}
	}
	c := New()
	if err := c.Init(&logic.Options{DBUrl: "sqlite://?mode=memory"}); err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	req := httptest.NewRequest("POST", "/mcp", nil)
	w := httptest.NewRecorder()
	c.APIHandler().ServeHTTP(w, req)
	if w.Code != 404 {
		t.Fatal("MCP enabled by default")
	}
}
func mustJSON(value interface{}) []byte { b, _ := json.Marshal(value); return b }

// Rejecting APNs at the no-device stage proves credentials passed validation
// without sending a real notification. Message construction is also checked below.
func TestMCPSendingAndIndependentRevocation(t *testing.T) {
	c, raw, id := mcpFixture(t, true, false)
	h := c.APIHandler()
	var imageData bytes.Buffer
	_ = png.Encode(&imageData, image.NewRGBA(image.Rect(0, 0, 1, 1)))
	for _, tc := range []struct {
		name string
		args map[string]interface{}
	}{
		{"send_text", map[string]interface{}{"channel": "development", "text": strings.Repeat("汉", 1200)}},
		{"send_text", map[string]interface{}{"channel": "sys.none", "text": "hello"}},
		{"send_image", map[string]interface{}{"channel": "development", "data_base64": base64.StdEncoding.EncodeToString(imageData.Bytes())}},
		{"send_file", map[string]interface{}{"channel": "development", "filename": "report.txt", "data_base64": "aGVsbG8="}},
	} {
		_, response := mcpRPC(t, h, raw, "tools/call", map[string]interface{}{"name": tc.name, "arguments": tc.args})
		body := string(mustJSON(response))
		if !strings.Contains(body, "no devices found") {
			t.Fatal(body)
		}
	}

	userKey := crypto.GenerateSecretKey([]byte("synthetic-mcp-account"))
	uid := userKey.ToID(0)
	device := crypto.GenerateSecretKey([]byte("synthetic-mcp-device"))
	hash := sha1.Sum(device.MarshalPublicKey())
	deviceID := strings.ToUpper(hex.EncodeToString(hash[:]))
	if err := c.logic.BindDevice(uid, deviceID, device.EncodePublicKey(), 0); err != nil {
		t.Fatal(err)
	}
	if err := c.logic.UpdatePushToken(uid, deviceID, crypto.Base64Encode.EncodeToString([]byte("synthetic-push")), false); err != nil {
		t.Fatal(err)
	}
	previous := logic.MockPusher
	logic.MockPusher = &MockAPNSPusher{}
	defer func() { logic.MockPusher = previous }()
	_, submitted := mcpRPC(t, h, raw, "tools/call", map[string]interface{}{"name": "send_text", "arguments": map[string]interface{}{"channel": "development", "text": strings.Repeat("汉", 1200)}})
	if !strings.Contains(string(mustJSON(submitted)), "text_attachment") || !strings.Contains(string(mustJSON(submitted)), "submitted") {
		t.Fatal(submitted)
	}
	auth, err := c.logic.AuthenticateMCP(raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.logic.UpdateMCPToken(auth.UserID, id, "", true); err != nil {
		t.Fatal(err)
	}
	if status, _ := mcpRPC(t, h, raw, "tools/list", map[string]interface{}{}); status != 401 {
		t.Fatal("revoked credential accepted")
	}
}

type mcpStreamRecorder struct {
	*httptest.ResponseRecorder
	flushed chan struct{}
	once    sync.Once
}

func (w *mcpStreamRecorder) Flush() {
	w.ResponseRecorder.Flush()
	w.once.Do(func() { close(w.flushed) })
}
func TestMCPListenRevocation(t *testing.T) {
	c, raw, id := mcpFixture(t, false, true)
	body := `{"jsonrpc":"2.0","id":9,"method":"subscriptions/listen","params":{"notifications":{"toolsListChanged":true},"_meta":{"io.modelcontextprotocol/protocolVersion":"2026-07-28","io.modelcontextprotocol/clientInfo":{"name":"test","version":"1"},"io.modelcontextprotocol/clientCapabilities":{}}}}`
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req := httptest.NewRequest("POST", "http://localhost/mcp", strings.NewReader(body)).WithContext(ctx)
	req.Header.Set("Authorization", "Bearer "+raw)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("MCP-Protocol-Version", "2026-07-28")
	req.Header.Set("Mcp-Method", "subscriptions/listen")
	w := &mcpStreamRecorder{ResponseRecorder: httptest.NewRecorder(), flushed: make(chan struct{})}
	done := make(chan struct{})
	go func() { defer close(done); c.APIHandler().ServeHTTP(w, req) }()
	select {
	case <-w.flushed:
	case <-done:
		t.Fatalf("listen ended: %d %s", w.Code, w.Body.String())
	case <-time.After(3 * time.Second):
		t.Fatal("listen did not open")
	}
	auth, err := c.logic.AuthenticateMCP(raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.logic.UpdateMCPToken(auth.UserID, id, "", true); err != nil {
		t.Fatal(err)
	}
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		cancel()
		<-done
		t.Fatal("revoked listen did not close")
	}
	if !strings.Contains(w.Header().Get("Content-Type"), "text/event-stream") {
		t.Fatal(w.Body.String())
	}
}

type mcpMemoryTransport struct {
	handler http.Handler
	token   string
}

func (t mcpMemoryTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	request = request.Clone(request.Context())
	request.Header = request.Header.Clone()
	request.Header.Set("Authorization", "Bearer "+t.token)
	w := httptest.NewRecorder()
	t.handler.ServeHTTP(w, request)
	return w.Result(), nil
}
func TestMCPRealSDKClient(t *testing.T) {
	c, raw, _ := mcpFixture(t, false, false)
	client := mcp.NewClient(&mcp.Implementation{Name: "integration-test", Version: "1"}, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: "http://localhost/mcp", HTTPClient: &http.Client{Transport: mcpMemoryTransport{handler: c.APIHandler(), token: raw}}, DisableStandaloneSSE: true}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = session.Close() }()
	tools, err := session.ListTools(ctx, nil)
	if err != nil || len(tools.Tools) != 1 {
		t.Fatal(tools, err)
	}
	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "send_text", Arguments: map[string]interface{}{"text": "hello"}})
	if err != nil || !result.IsError || !strings.Contains(string(mustJSON(result)), "no devices found") {
		t.Fatal(result, err)
	}
}

func TestMCPStreamHeartbeat(t *testing.T) {
	for _, ctype := range []string{"application/json", "text/event-stream"} {
		recorder := httptest.NewRecorder()
		writer := &mcpStreamWriter{ResponseWriter: recorder}
		writer.Header().Set("Content-Type", ctype)
		writer.WriteHeader(200)
		writer.heartbeat()
		if strings.Contains(recorder.Body.String(), "keepalive") != (ctype == "text/event-stream") {
			t.Fatal("heartbeat corrupted response format")
		}
	}
}
func TestMCPListenDisabled(t *testing.T) {
	c, raw, _ := mcpFixture(t, false, false)
	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer "+raw)
	req.Header.Set("Mcp-Method", "subscriptions/listen")
	w := httptest.NewRecorder()
	c.APIHandler().ServeHTTP(w, req)
	if w.Code != 405 {
		t.Fatal("SSE enabled without opt in")
	}
}

func TestMCPFeatureFiltering(t *testing.T) {
	c, raw, _ := mcpFixture(t, true, false)
	var features []string
	for _, feature := range c.logic.Features {
		if feature != "msg.image" && feature != "msg.file" {
			features = append(features, feature)
		}
	}
	c.logic.Features = features
	_, response := mcpRPC(t, c.APIHandler(), raw, "tools/list", map[string]interface{}{})
	encoded := string(mustJSON(response))
	if strings.Contains(encoded, "send_image") || strings.Contains(encoded, "send_file") || !strings.Contains(encoded, `400`) {
		t.Fatal("tool discovery ignored effective capabilities")
	}
}
func TestMCPCredentialManagementAuthorization(t *testing.T) {
	c, _, _ := mcpFixture(t, false, false)
	h := c.APIHandler()
	key := crypto.GenerateSecretKey([]byte("synthetic-mcp-account"))
	invoke := func(signed bool, params map[string]interface{}) (int, map[string]interface{}) {
		params["user"] = key.ToID(0)
		params["nonce"] = time.Now().UnixMilli()
		body := mustJSON(params)
		req := httptest.NewRequest("POST", "/rest/v1/mcp-tokens", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if signed {
			sign, _ := key.Sign(body)
			req.Header.Set("CHUserSign", crypto.Base64Encode.EncodeToString(sign))
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		var result map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &result)
		return w.Code, result
	}
	if status, _ := invoke(false, map[string]interface{}{"action": "list"}); status != 401 {
		t.Fatal("unsigned credential administration accepted")
	}
	status, result := invoke(true, map[string]interface{}{"action": "create", "name": "Independent agent", "expires": time.Now().Add(time.Hour).Unix()})
	if status != 200 {
		t.Fatal(result)
	}
	raw, ok := result["token"].(string)
	if !ok || !strings.HasPrefix(raw, "cmcp1.") {
		t.Fatal("missing one-time credential")
	}
	status, result = invoke(true, map[string]interface{}{"action": "list"})
	if status != 200 || strings.Contains(string(mustJSON(result)), raw) {
		t.Fatal("list exposed bearer")
	}
}
