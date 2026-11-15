package core

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/png"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chanify/chanify/crypto"
	"github.com/chanify/chanify/logic"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func signedManagement(t *testing.T, c *Core, path string, p map[string]interface{}) *httptest.ResponseRecorder {
	t.Helper()
	key := crypto.GenerateSecretKey([]byte("synthetic-mcp-account"))
	p["user"] = key.ToID(0)
	p["nonce"] = time.Now().UnixMilli()
	body := mustJSON(p)
	signature, err := key.Sign(body)
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest("POST", path, bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("CHUserSign", crypto.Base64Encode.EncodeToString(signature))
	w := httptest.NewRecorder()
	c.APIHandler().ServeHTTP(w, r)
	return w
}

func TestManagementErrorResponses(t *testing.T) {
	c, _, id := mcpFixture(t, false, false)
	for _, path := range []string{"/rest/v1/mcp-tokens", "/rest/v1/server-tokens", "/rest/v1/delete-user"} {
		r := httptest.NewRequest("POST", path, strings.NewReader("{"))
		w := httptest.NewRecorder()
		c.APIHandler().ServeHTTP(w, r)
		if w.Code != 400 {
			t.Fatal(path, w.Code)
		}
	}
	for _, p := range []map[string]interface{}{{"action": "create", "name": "bad", "expires": 1}, {"action": "rename", "id": "missing", "name": "name"}, {"action": "unknown"}} {
		if w := signedManagement(t, c, "/rest/v1/mcp-tokens", p); w.Code != 400 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	for _, action := range []string{"rename", "revoke"} {
		if w := signedManagement(t, c, "/rest/v1/mcp-tokens", map[string]interface{}{"action": action, "id": id, "name": "Renamed"}); w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	if w := signedManagement(t, c, "/rest/v1/server-tokens", map[string]interface{}{"action": "rename", "id": strings.Repeat("1", 32), "name": "Name"}); w.Code != 404 {
		t.Fatal(w.Code, w.Body.String())
	}
}

func TestMCPResultAdaptersAndMissingAuthorization(t *testing.T) {
	r := &mcpSendResult{}
	r.DataFromReader(400, 0, "", strings.NewReader(`{"msg":"rejected"}`), nil)
	if r.code != 400 || r.message() != "rejected" {
		t.Fatal(r)
	}
	r.data = nil
	if r.message() != "notification submission failed" {
		t.Fatal(r.message())
	}
	w := httptest.NewRecorder()
	stream := &mcpStreamWriter{ResponseWriter: w}
	if stream.Unwrap() != w {
		t.Fatal("underlying writer lost")
	}
	c, _, _ := mcpFixture(t, false, false)
	result, err := c.mcpTool("send_text")(context.Background(), &mcp.CallToolRequest{})
	if err != nil || !result.IsError {
		t.Fatal(result, err)
	}
}

func TestCoreInitFailures(t *testing.T) {
	if err := New().Init(&logic.Options{AdminPassword: "short"}); err == nil {
		t.Fatal("invalid password accepted")
	}
	c := New()
	if err := c.Init(&logic.Options{DBUrl: "sqlite://?mode=memory", AdminPassword: "synthetic-admin-password", Endpoint: "https://node.example"}); err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if c.admin == nil {
		t.Fatal("admin not enabled")
	}
}

func TestMCPToolRejectsRevokedAuthorization(t *testing.T) {
	c, raw, _ := mcpFixture(t, true, false)
	auth := &mcpAuthorization{raw: raw}
	ctx := context.WithValue(context.Background(), mcpAuthKey{}, auth)
	result, err := c.mcpTool("send_text")(ctx, &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Arguments: json.RawMessage(`{"text":"hello"}`)}})
	if err != nil || !result.IsError {
		t.Fatal(result, err)
	}
	// A revoked credential is rejected on the direct handler path too.
	records, err := c.logic.ListMCPTokens(crypto.GenerateSecretKey([]byte("synthetic-mcp-account")).ToID(0))
	if err != nil {
		t.Fatal(err)
	}
	if err := c.logic.UpdateMCPToken(records[0].UserID, records[0].ID, "", true); err != nil {
		t.Fatal(err)
	}
	result, err = c.mcpTool("send_text")(ctx, &mcp.CallToolRequest{})
	if err != nil || !result.IsError {
		t.Fatal(result, err)
	}
}

func TestMCPAttachmentStorageFailures(t *testing.T) {
	root := t.TempDir()
	c := New()
	if err := c.Init(&logic.Options{DBUrl: "sqlite://?mode=memory", FilePath: root, Registerable: true, MCPEnabled: true}); err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	key := crypto.GenerateSecretKey([]byte("synthetic-storage-failure"))
	uid := key.ToID(0)
	if _, err := c.logic.UpsertUser(uid, key.EncodePublicKey(), false); err != nil {
		t.Fatal(err)
	}
	_, raw, err := c.logic.CreateMCPToken(uid, "Agent", time.Now().Add(time.Hour).Unix())
	if err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{"files", "images"} {
		if err := os.Remove(filepath.Join(root, dir)); err != nil {
			t.Fatal(err)
		}
	}
	var imageData bytes.Buffer
	if err := png.Encode(&imageData, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		args map[string]interface{}
	}{
		{"send_text", map[string]interface{}{"text": strings.Repeat("long report ", 200)}},
		{"send_file", map[string]interface{}{"data_base64": "YQ==", "filename": "report.txt"}},
		{"send_image", map[string]interface{}{"data_base64": base64.StdEncoding.EncodeToString(imageData.Bytes())}},
	} {
		ctx := context.WithValue(context.Background(), mcpAuthKey{}, &mcpAuthorization{raw: raw})
		result, err := c.mcpTool(tc.name)(ctx, &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Arguments: mustJSON(tc.args)}})
		if err != nil || !result.IsError {
			t.Fatal(result, err)
		}
		data, ok := result.StructuredContent.(map[string]interface{})
		if !ok || data["code"] != "storage_failed" {
			t.Fatal(result)
		}
	}
}
