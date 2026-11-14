package core

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"github.com/chanify/chanify/crypto"
	"github.com/chanify/chanify/pb"
	"google.golang.org/protobuf/proto"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/chanify/chanify/logic"
	"golang.org/x/crypto/bcrypt"
)

var adminTestHash string
var adminTestOnce sync.Once

func testAdminHash(t *testing.T) string {
	t.Helper()
	adminTestOnce.Do(func() {
		hash, err := bcrypt.GenerateFromPassword([]byte(adminDigest("synthetic-admin-password")), 10)
		if err != nil {
			t.Fatal(err)
		}
		adminTestHash = string(hash)
	})
	return adminTestHash
}
func adminRequest(h http.Handler, method, path string, body interface{}, cookie *http.Cookie, csrf string) *httptest.ResponseRecorder {
	data, _ := json.Marshal(body)
	r := httptest.NewRequest(method, "http://localhost"+path, bytes.NewReader(data))
	r.Header.Set("Content-Type", "application/json")
	if cookie != nil {
		r.AddCookie(cookie)
	}
	if csrf != "" {
		r.Header.Set("X-CSRF-Token", csrf)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}
func adminLoginTest(t *testing.T, h http.Handler) (*http.Cookie, string) {
	t.Helper()
	w := adminRequest(h, "POST", "/admin/api/login", map[string]string{"password_digest": adminDigest("synthetic-admin-password")}, nil, "")
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var session struct {
		CSRF string `json:"csrf"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	cookies := w.Result().Cookies()
	if len(cookies) != 1 || session.CSRF == "" {
		t.Fatal("missing session")
	}
	return cookies[0], session.CSRF
}
func TestAdminDisabledAndConfiguration(t *testing.T) {
	c := New()
	if err := c.Init(&logic.Options{DBUrl: "sqlite://?mode=memory"}); err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	for _, path := range []string{"/admin", "/admin/api/state", "/admin/api/login"} {
		if w := adminRequest(c.APIHandler(), "GET", path, nil, nil, ""); w.Code != 404 {
			t.Fatal("disabled admin route exposed", path, w.Code)
		}
	}
	if err := New().Init(&logic.Options{AdminPassword: "synthetic-admin-password"}); err == nil {
		t.Fatal("nonpersistent admin accepted")
	}
}
func TestAdminAuthenticationAndAssets(t *testing.T) {
	c, _, _ := mcpFixture(t, false, false)
	c.admin = newAdmin(testAdminHash(t), "https://localhost/push")
	h := c.APIHandler()
	for _, path := range []string{"/admin", "/admin/app.js", "/admin/style.css"} {
		w := adminRequest(h, "GET", path, nil, nil, "")
		if w.Code != 200 || w.Body.Len() == 0 || w.Header().Get("Cache-Control") != "no-store" || w.Header().Get("Content-Security-Policy") == "" {
			t.Fatal(path, w.Code)
		}
	}
	if w := adminRequest(h, "GET", "/admin/api/state", nil, nil, ""); w.Code != 401 {
		t.Fatal(w.Code)
	}
	cookie, csrf := adminLoginTest(t, h)
	if !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteStrictMode || cookie.Path != "/push/admin" {
		t.Fatal("unsafe cookie", cookie)
	}
	if w := adminRequest(h, "POST", "/admin/api/logout", map[string]string{}, cookie, ""); w.Code != 403 {
		t.Fatal("missing CSRF accepted", w.Code)
	}
	req := httptest.NewRequest("GET", "http://localhost/admin/api/state", nil)
	req.AddCookie(cookie)
	req.Header.Set("Origin", "https://evil.example")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != 403 {
		t.Fatal("cross origin accepted")
	}
	if w := adminRequest(h, "GET", "/admin/api/state", nil, cookie, ""); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if w := adminRequest(h, "POST", "/admin/api/logout", map[string]string{}, cookie, csrf); w.Code != 200 {
		t.Fatal(w.Code)
	}
	if w := adminRequest(h, "GET", "/admin/api/session", nil, cookie, ""); w.Code != 401 {
		t.Fatal("logout did not invalidate")
	}
	cookie, _ = adminLoginTest(t, h)
	c.admin = newAdmin(testAdminHash(t), "")
	if w := adminRequest(h, "GET", "/admin/api/session", nil, cookie, ""); w.Code != 401 {
		t.Fatal("session survived restart")
	}
}
func TestAdminTokenLifecycleAndPrivacy(t *testing.T) {
	c, raw, id := mcpFixture(t, false, false)
	c.admin = newAdmin(testAdminHash(t), "")
	h := c.APIHandler()
	record, err := c.logic.AuthenticateMCP(raw)
	if err != nil {
		t.Fatal(err)
	}
	cookie, csrf := adminLoginTest(t, h)
	w := adminRequest(h, "GET", "/admin/api/state", nil, cookie, "")
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	for _, secret := range []string{raw, `"hash"`, `"encrypted_token"`, `"secret_key"`, `"public_key"`} {
		if strings.Contains(w.Body.String(), secret) {
			t.Fatal("private field leaked", secret)
		}
	}
	var snapshot struct {
		Tokens []struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			Revoked bool   `json:"revoked"`
		} `json:"tokens"`
		Users []interface{} `json:"users"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &snapshot); err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Tokens) != 1 || snapshot.Tokens[0].ID != id || len(snapshot.Users) != 1 {
		t.Fatal(w.Body.String())
	}
	action := map[string]interface{}{"action": "rename", "kind": "mcp", "user": record.UserID, "id": id, "name": "New Agent"}
	if w := adminRequest(h, "POST", "/admin/api/tokens", action, cookie, csrf); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if authenticated, err := c.logic.AuthenticateMCP(raw); err != nil || authenticated.Name != "New Agent" {
		t.Fatal("rename altered authentication", err)
	}
	action["action"] = "revoke"
	if w := adminRequest(h, "POST", "/admin/api/tokens", action, cookie, csrf); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if status, _ := mcpRPC(t, h, raw, "tools/list", map[string]interface{}{}); status != 401 {
		t.Fatal("revoked bearer accepted", status)
	}
	create := map[string]interface{}{"action": "create", "kind": "mcp", "user": record.UserID, "name": "Web Agent", "expires": time.Now().Add(time.Hour).Unix()}
	w = adminRequest(h, "POST", "/admin/api/tokens", create, cookie, csrf)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var created struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil || created.Token == "" {
		t.Fatal("missing one-time token")
	}
	if _, err := c.logic.AuthenticateMCP(created.Token); err != nil {
		t.Fatal(err)
	}
	w = adminRequest(h, "GET", "/admin/api/state", nil, cookie, "")
	if w.Code != 200 || strings.Contains(w.Body.String(), created.Token) || strings.Contains(w.Body.String(), raw) {
		t.Fatal("token exposed in inventory")
	}
}
func TestAdminRateLimitExpiryAndActivity(t *testing.T) {
	c, raw, _ := mcpFixture(t, false, false)
	c.admin = newAdmin(testAdminHash(t), "")
	h := c.APIHandler()
	for i := 0; i < 5; i++ {
		if w := adminRequest(h, "POST", "/admin/api/login", map[string]string{"password_digest": adminDigest("wrong")}, nil, ""); w.Code != 401 {
			t.Fatal(w.Code)
		}
	}
	if w := adminRequest(h, "POST", "/admin/api/login", map[string]string{"password_digest": adminDigest("synthetic-admin-password")}, nil, ""); w.Code != 429 {
		t.Fatal("login limit missing", w.Code)
	}
	c.admin = newAdmin(testAdminHash(t), "")
	cookie, _ := adminLoginTest(t, h)
	c.admin.mu.Lock()
	session := c.admin.sessions[adminDigest(cookie.Value)]
	session.Expires = time.Now().Add(-time.Second)
	c.admin.sessions[adminDigest(cookie.Value)] = session
	c.admin.mu.Unlock()
	if w := adminRequest(h, "GET", "/admin/api/state", nil, cookie, ""); w.Code != 401 {
		t.Fatal("expired session accepted")
	}
	if status, _ := mcpRPC(t, h, raw, "tools/list", map[string]interface{}{}); status != 200 {
		t.Fatal(status)
	}
	if len(c.admin.clients) != 1 {
		t.Fatal("authenticated activity missing")
	}
	for _, client := range c.admin.clients {
		if client.ActiveRequests != 0 || client.Listeners != 0 || client.Kind != "mcp" {
			t.Fatal("short request shown as connected")
		}
	}
	r := httptest.NewRequest("GET", "http://localhost/mcp", nil)
	start, finish := c.beginAdminActivity(r, "mcp", "synthetic-id", "synthetic-user")
	start()
	start()
	c.admin.mu.Lock()
	listeners := 0
	for _, client := range c.admin.clients {
		listeners += client.Listeners
	}
	c.admin.mu.Unlock()
	if listeners != 1 {
		t.Fatal("listener count", listeners)
	}
	finish()
	for _, client := range c.admin.clients {
		if client.ActiveRequests != 0 || client.Listeners != 0 {
			t.Fatal("connection not released")
		}
	}
}

func TestAdminSendingTokenRevocation(t *testing.T) {
	c := New()
	if err := c.Init(&logic.Options{DBUrl: "sqlite://?mode=memory", Registerable: true, AdminPassword: "synthetic-admin-password"}); err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	key := crypto.GenerateSecretKey([]byte("synthetic-admin-sending-account"))
	uid := key.ToID(0)
	user, err := c.logic.UpsertUser(uid, key.EncodePublicKey(), false)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := proto.Marshal(&pb.Token{Scope: pb.TokenScope_ServerSend, TokenId: bytes.Repeat([]byte{3}, 16), UserId: uid, NodeId: c.logic.NodeID, Expires: uint64(time.Now().Add(time.Hour).Unix())})
	mac := hmac.New(sha256.New, user.SecretKey[:32])
	_, _ = mac.Write(body)
	raw := "cst1:" + crypto.Base64Encode.EncodeToString(body) + ".." + crypto.Base64Encode.EncodeToString(mac.Sum(nil))
	records, err := c.logic.ManageServerTokens(uid, "create", "", "Phone", raw)
	if err != nil {
		t.Fatal(err)
	}
	h := c.APIHandler()
	cookie, csrf := adminLoginTest(t, h)
	w := adminRequest(h, "GET", "/admin/api/state", nil, cookie, "")
	if w.Code != 200 || strings.Contains(w.Body.String(), raw) || strings.Contains(w.Body.String(), records[0].EncryptedToken) || strings.Contains(w.Body.String(), records[0].Hash) {
		t.Fatal("sending token material exposed")
	}
	if _, err := c.parseToken(raw); err != nil {
		t.Fatal(err)
	}
	// A sending credential is not an administrator session.
	req := httptest.NewRequest("GET", "http://localhost/admin/api/state", nil)
	req.Header.Set("Authorization", "Bearer "+raw)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != 401 {
		t.Fatal("sending token gained administration")
	}
	action := map[string]string{"action": "rename", "kind": "sending", "user": uid, "id": records[0].ID, "name": "Renamed phone"}
	if w := adminRequest(h, "POST", "/admin/api/tokens", action, cookie, csrf); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if _, err := c.parseToken(raw); err != nil {
		t.Fatal("rename broke existing sending token", err)
	}
	action["action"] = "revoke"
	if w := adminRequest(h, "POST", "/admin/api/tokens", action, cookie, csrf); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if _, err := c.parseToken(raw); err == nil {
		t.Fatal("revoked sending token accepted")
	}
	if w := adminRequest(h, "GET", "/admin/api/state", nil, cookie, ""); w.Code != 200 || !strings.Contains(w.Body.String(), `"revoked":true`) {
		t.Fatal("revocation missing from inventory")
	}
}

func TestAdminRevocationClosesMCPListener(t *testing.T) {
	c, raw, id := mcpFixture(t, false, true)
	c.admin = newAdmin(testAdminHash(t), "")
	h := c.APIHandler()
	cookie, csrf := adminLoginTest(t, h)
	auth, err := c.logic.AuthenticateMCP(raw)
	if err != nil {
		t.Fatal(err)
	}
	body := `{"jsonrpc":"2.0","id":9,"method":"subscriptions/listen","params":{"notifications":{"toolsListChanged":true},"_meta":{"io.modelcontextprotocol/protocolVersion":"2026-07-28","io.modelcontextprotocol/clientInfo":{"name":"synthetic-agent","version":"1"},"io.modelcontextprotocol/clientCapabilities":{}}}}`
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
	go func() { defer close(done); h.ServeHTTP(w, req) }()
	select {
	case <-w.flushed:
	case <-done:
		t.Fatal("listener did not open")
	case <-time.After(3 * time.Second):
		t.Fatal("listener timed out")
	}
	c.admin.mu.Lock()
	listeners := 0
	for _, client := range c.admin.clients {
		listeners += client.Listeners
		if client.ConnectedSince == 0 {
			t.Error("missing connection time")
		}
	}
	c.admin.mu.Unlock()
	if listeners != 1 {
		t.Fatal("listener activity not recorded", listeners)
	}
	action := map[string]string{"action": "revoke", "kind": "mcp", "user": auth.UserID, "id": id}
	if response := adminRequest(h, "POST", "/admin/api/tokens", action, cookie, csrf); response.Code != 200 {
		t.Fatal(response.Code, response.Body.String())
	}
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		cancel()
		<-done
		t.Fatal("web revocation did not close listener")
	}
	c.admin.mu.Lock()
	defer c.admin.mu.Unlock()
	for _, client := range c.admin.clients {
		if client.Listeners != 0 || client.ActiveRequests != 0 || client.ConnectedSince != 0 {
			t.Fatal("closed listener still counted")
		}
	}
}

func TestAdminPlainPasswordConfiguration(t *testing.T) {
	const password = "synthetic-admin-password"
	c := New()
	if err := c.Init(&logic.Options{DBUrl: "sqlite://?mode=memory", AdminPassword: password}); err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if bytes.Contains(c.admin.passwordHash, []byte(password)) {
		t.Fatal("plain password retained in login verifier")
	}
	cookie, _ := adminLoginTest(t, c.APIHandler())
	if w := adminRequest(c.APIHandler(), "GET", "/admin/api/state", nil, cookie, ""); w.Code != 200 {
		t.Fatal(w.Code)
	}
	for _, value := range []string{"too-short", strings.Repeat("x", 73), strings.Repeat("x", 12) + string([]byte{0xff})} {
		node := New()
		err := node.Init(&logic.Options{DBUrl: "sqlite://?mode=memory", AdminPassword: value})
		if err == nil {
			node.Close()
			t.Fatal("invalid password accepted")
		}
		if strings.Contains(err.Error(), value) {
			t.Fatal("password exposed in configuration error")
		}
	}
	// Empty values preserve the disabled-by-default behavior.
	disabled := New()
	if err := disabled.Init(&logic.Options{DBUrl: "sqlite://?mode=memory", AdminPassword: ""}); err != nil {
		t.Fatal(err)
	}
	defer disabled.Close()
	if w := adminRequest(disabled.APIHandler(), "GET", "/admin", nil, nil, ""); w.Code != 404 {
		t.Fatal("empty password enabled administration")
	}
}

func TestAdminLoginRequiresPasswordDigest(t *testing.T) {
	c := New()
	password := "管理密码-synthetic-123"
	if err := c.Init(&logic.Options{DBUrl: "sqlite://?mode=memory", AdminPassword: password}); err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	h := c.APIHandler()
	for _, body := range []map[string]string{
		{"password": password},
		{"password_digest": password},
		{"password_digest": strings.Repeat("z", 64)},
		{"password_digest": ""},
	} {
		if w := adminRequest(h, "POST", "/admin/api/login", body, nil, ""); w.Code != 400 {
			t.Fatal("non-digest login accepted", w.Code)
		}
	}
	if w := adminRequest(h, "POST", "/admin/api/login", map[string]string{"password_digest": adminDigest("wrong")}, nil, ""); w.Code != 401 {
		t.Fatal("wrong digest accepted", w.Code)
	}
	if w := adminRequest(h, "POST", "/admin/api/login", map[string]string{"password_digest": adminDigest(password)}, nil, ""); w.Code != 200 {
		t.Fatal("UTF-8 password digest rejected", w.Code)
	}
}

func TestAdminTrailingSlashPreservesProxyPrefix(t *testing.T) {
	c, _, _ := mcpFixture(t, false, false)
	c.admin = newAdmin(testAdminHash(t), "https://node.example/prefix")
	handler := c.APIHandler()
	for _, prefix := range []string{"", "/prefix", "/nested/prefix", "/prefix%20space"} {
		t.Run(prefix, func(t *testing.T) {
			external, err := url.Parse("https://node.example" + prefix + "/admin/")
			if err != nil {
				t.Fatal(err)
			}
			// The reverse proxy has removed the prefix before forwarding the request.
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, httptest.NewRequest("GET", "https://node.example/admin/", nil))
			if w.Code != http.StatusTemporaryRedirect {
				t.Fatal("expected temporary redirect", w.Code)
			}
			location, err := url.Parse(w.Header().Get("Location"))
			if err != nil {
				t.Fatal(err)
			}
			if resolved := external.ResolveReference(location).String(); resolved != "https://node.example"+prefix+"/admin" {
				t.Fatalf("redirect lost external prefix: %s", resolved)
			}
			follow := httptest.NewRecorder()
			handler.ServeHTTP(follow, httptest.NewRequest("GET", "https://node.example/admin", nil))
			if follow.Code != http.StatusOK {
				t.Fatal("redirect destination unavailable", follow.Code)
			}
		})
	}
}
