package core

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"embed"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

//go:embed adminassets/*
var adminAssets embed.FS

const adminCookie = "chanify_admin"
const adminLifetime = 8 * time.Hour

type adminSession struct {
	CSRF    string
	Expires time.Time
}
type adminAttempt struct {
	Count int
	Until time.Time
}
type adminClient struct {
	ID             string `json:"id"`
	Kind           string `json:"kind"`
	CredentialID   string `json:"credential_id"`
	UserID         string `json:"user"`
	IP             string `json:"ip"`
	UserAgent      string `json:"user_agent"`
	LastSeen       int64  `json:"last_seen"`
	ActiveRequests int    `json:"active_requests"`
	Listeners      int    `json:"listeners"`
	ConnectedSince int64  `json:"connected_since,omitempty"`
}
type adminState struct {
	mu           sync.Mutex
	passwordHash []byte
	sessions     map[string]adminSession
	attempts     map[string]adminAttempt
	clients      map[string]*adminClient
	loginSlots   chan struct{}
	cookiePath   string
	secure       bool
}

// Derive the login verifier once at startup; never retain the plain password here.
func adminPasswordHash(password string) (string, error) {
	if password == "" {
		return "", nil
	}
	if len(password) < 12 || len(password) > 72 || !utf8.ValidString(password) {
		return "", errors.New("server.web_admin.password must contain 12 through 72 UTF-8 bytes")
	}
	data, err := bcrypt.GenerateFromPassword([]byte(adminDigest(password)), 12)
	return string(data), err
}

func newAdmin(hash, endpoint string) *adminState {
	u, _ := url.Parse(endpoint)
	path := "/admin"
	secure := false
	if u != nil {
		path = strings.TrimRight(u.Path, "/") + "/admin"
		secure = u.Scheme == "https"
	}
	return &adminState{passwordHash: []byte(hash), sessions: map[string]adminSession{}, attempts: map[string]adminAttempt{}, clients: map[string]*adminClient{}, loginSlots: make(chan struct{}, 2), cookiePath: path, secure: secure}
}
func adminRandom() (string, error) {
	data := make([]byte, 32)
	if _, err := rand.Read(data); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}
func adminDigest(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}
func (r *adminState) session(request *http.Request) (adminSession, bool) {
	cookie, err := request.Cookie(adminCookie)
	if err != nil {
		return adminSession{}, false
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	key := adminDigest(cookie.Value)
	session, ok := r.sessions[key]
	if ok && !time.Now().Before(session.Expires) {
		delete(r.sessions, key)
		return adminSession{}, false
	}
	return session, ok
}
func adminJSON(w http.ResponseWriter, code int, value interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(value)
}
func adminDecode(w http.ResponseWriter, r *http.Request, value interface{}) error {
	if strings.Split(r.Header.Get("Content-Type"), ";")[0] != "application/json" {
		return errors.New("JSON required")
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	if decoder.Decode(new(interface{})) != io.EOF {
		return errors.New("one JSON object required")
	}
	return nil
}
func (c *Core) adminHandler() http.Handler {
	mux := http.NewServeMux()
	for _, asset := range []struct{ path, name, mime string }{{"/admin", "index.html", "text/html"}, {"/admin/app.js", "app.js", "text/javascript"}, {"/admin/style.css", "style.css", "text/css"}, {"/admin/preferences.js", "preferences.js", "text/javascript"}, {"/admin/logo.png", "logo.png", "image/png"}} {
		mux.HandleFunc("GET "+asset.path, func(w http.ResponseWriter, r *http.Request) {
			data, err := adminAssets.ReadFile("adminassets/" + asset.name)
			if err != nil {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", asset.mime+"; charset=utf-8")
			_, _ = w.Write(data)
		})
	}
	mux.HandleFunc("GET /admin/{$}", func(w http.ResponseWriter, r *http.Request) {
		// Preserve the browser's external prefix when a proxy strips it upstream.
		w.Header().Set("Location", "../admin")
		w.WriteHeader(http.StatusTemporaryRedirect)
	})
	mux.HandleFunc("POST /admin/api/login", c.adminLogin)
	mux.HandleFunc("GET /admin/api/session", func(w http.ResponseWriter, r *http.Request) {
		session, ok := c.admin.session(r)
		if !ok {
			adminJSON(w, 401, map[string]string{"error": "请先登录"})
			return
		}
		adminJSON(w, 200, map[string]interface{}{"csrf": session.CSRF, "expires": session.Expires.Unix()})
	})
	mux.Handle("POST /admin/api/logout", c.adminAuthorized(func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie(adminCookie); err == nil {
			c.admin.mu.Lock()
			delete(c.admin.sessions, adminDigest(cookie.Value))
			c.admin.mu.Unlock()
		}
		c.adminCookie(w, r, "", -1)
		adminJSON(w, 200, map[string]bool{"success": true})
	}))
	mux.Handle("GET /admin/api/state", c.adminAuthorized(c.adminSnapshot))
	mux.Handle("POST /admin/api/tokens", c.adminAuthorized(c.adminTokens))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		if origin := r.Header.Get("Origin"); origin != "" {
			u, err := url.Parse(origin)
			if err != nil || u.Host != r.Host || (u.Scheme != "http" && u.Scheme != "https") {
				adminJSON(w, 403, map[string]string{"error": "请求来源无效"})
				return
			}
		}
		mux.ServeHTTP(w, r)
	})
}
func (c *Core) adminCookie(w http.ResponseWriter, r *http.Request, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{Name: adminCookie, Value: value, Path: c.admin.cookiePath, HttpOnly: true, Secure: c.admin.secure || r.TLS != nil, SameSite: http.SameSiteStrictMode, MaxAge: maxAge})
}
func (c *Core) adminLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PasswordDigest string `json:"password_digest"`
	}
	if adminDecode(w, r, &body) != nil || len(body.PasswordDigest) != 64 {
		adminJSON(w, 400, map[string]string{"error": "请输入有效密码"})
		return
	}
	if _, err := hex.DecodeString(body.PasswordDigest); err != nil {
		adminJSON(w, 400, map[string]string{"error": "请输入有效密码"})
		return
	}
	body.PasswordDigest = strings.ToLower(body.PasswordDigest)
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr
	}
	now := time.Now()
	state := c.admin
	state.mu.Lock()
	for key, attempt := range state.attempts {
		if now.After(attempt.Until) {
			delete(state.attempts, key)
		}
	}
	attempt, exists := state.attempts[ip]
	if (exists && attempt.Count >= 5) || (!exists && len(state.attempts) >= 1024) {
		state.mu.Unlock()
		w.Header().Set("Retry-After", "300")
		adminJSON(w, 429, map[string]string{"error": "尝试次数过多，请稍后重试"})
		return
	}
	if !exists {
		attempt.Until = now.Add(5 * time.Minute)
	}
	attempt.Count++
	state.attempts[ip] = attempt
	state.mu.Unlock()
	select {
	case state.loginSlots <- struct{}{}:
		defer func() { <-state.loginSlots }()
	default:
		adminJSON(w, 429, map[string]string{"error": "登录繁忙，请稍后重试"})
		return
	}
	if bcrypt.CompareHashAndPassword(state.passwordHash, []byte(body.PasswordDigest)) != nil {
		adminJSON(w, 401, map[string]string{"error": "密码不正确"})
		return
	}
	token, err := adminRandom()
	if err != nil {
		adminJSON(w, 500, map[string]string{"error": "无法创建会话"})
		return
	}
	csrf, err := adminRandom()
	if err != nil {
		adminJSON(w, 500, map[string]string{"error": "无法创建会话"})
		return
	}
	session := adminSession{CSRF: csrf, Expires: now.Add(adminLifetime)}
	state.mu.Lock()
	for key, old := range state.sessions {
		if now.After(old.Expires) {
			delete(state.sessions, key)
		}
	}
	if cookie, err := r.Cookie(adminCookie); err == nil {
		delete(state.sessions, adminDigest(cookie.Value))
	}
	if len(state.sessions) >= 256 {
		state.mu.Unlock()
		adminJSON(w, 429, map[string]string{"error": "会话数量已达上限"})
		return
	}
	delete(state.attempts, ip)
	state.sessions[adminDigest(token)] = session
	state.mu.Unlock()
	c.adminCookie(w, r, token, int(adminLifetime.Seconds()))
	adminJSON(w, 200, map[string]interface{}{"csrf": csrf, "expires": session.Expires.Unix()})
}
func (c *Core) adminAuthorized(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, ok := c.admin.session(r)
		if !ok {
			adminJSON(w, 401, map[string]string{"error": "登录已过期，请重新登录"})
			return
		}
		if r.Method != "GET" && subtle.ConstantTimeCompare([]byte(session.CSRF), []byte(r.Header.Get("X-CSRF-Token"))) != 1 {
			adminJSON(w, 403, map[string]string{"error": "请求校验失败，请刷新页面"})
			return
		}
		next(w, r)
	})
}
func (c *Core) adminSnapshot(w http.ResponseWriter, r *http.Request) {
	users, tokens, err := c.logic.AdminRecords()
	if err != nil {
		adminJSON(w, 500, map[string]string{"error": "无法读取管理数据"})
		return
	}
	state := c.admin
	state.mu.Lock()
	clients := []adminClient{}
	for _, client := range state.clients {
		clients = append(clients, *client)
	}
	state.mu.Unlock()
	sort.Slice(clients, func(i, j int) bool { return clients[i].LastSeen > clients[j].LastSeen })
	adminJSON(w, 200, map[string]interface{}{"node": map[string]interface{}{"name": c.logic.Name, "version": c.logic.Version, "features": c.logic.Features}, "users": users, "tokens": tokens, "clients": clients, "now": time.Now().Unix()})
}
func (c *Core) adminTokens(w http.ResponseWriter, r *http.Request) {
	var p struct {
		Action  string `json:"action"`
		Kind    string `json:"kind"`
		User    string `json:"user"`
		ID      string `json:"id"`
		Name    string `json:"name"`
		Expires int64  `json:"expires"`
	}
	if adminDecode(w, r, &p) != nil {
		adminJSON(w, 400, map[string]string{"error": "参数无效"})
		return
	}
	if p.Action == "create" && p.Kind == "mcp" {
		record, raw, err := c.logic.CreateMCPToken(p.User, p.Name, p.Expires)
		if err != nil {
			adminJSON(w, 400, map[string]string{"error": "创建失败，请检查用户、名称和有效期"})
			return
		}
		adminJSON(w, 200, map[string]interface{}{"credential": record, "token": raw})
		return
	}
	if p.Action != "rename" && p.Action != "revoke" {
		adminJSON(w, 400, map[string]string{"error": "不支持此操作"})
		return
	}
	if err := c.logic.AdminUpdateToken(p.Kind, p.User, p.ID, p.Name, p.Action == "revoke"); err != nil {
		adminJSON(w, 400, map[string]string{"error": "操作失败，令牌可能已失效"})
		return
	}
	adminJSON(w, 200, map[string]bool{"success": true})
}

// Activity is bounded, ephemeral and records only authenticated requests.
func (c *Core) beginAdminActivity(r *http.Request, kind, id, uid string) (func(), func()) {
	if c.admin == nil || (kind == "sending" && id == "") {
		return func() {}, func() {}
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr
	}
	agent := r.UserAgent()
	if len(agent) > 256 {
		agent = agent[:256]
	}
	key := adminDigest(fmt.Sprintf("%s/%s/%s/%s/%s", kind, uid, id, ip, agent))
	state := c.admin
	state.mu.Lock()
	if state.clients[key] == nil && len(state.clients) >= 2048 {
		oldest := ""
		var ts int64
		for k, v := range state.clients {
			if v.ActiveRequests == 0 && (oldest == "" || v.LastSeen < ts) {
				oldest = k
				ts = v.LastSeen
			}
		}
		if oldest == "" {
			state.mu.Unlock()
			return func() {}, func() {}
		}
		delete(state.clients, oldest)
	}
	client := state.clients[key]
	if client == nil {
		client = &adminClient{ID: key, Kind: kind, CredentialID: id, UserID: uid, IP: ip, UserAgent: agent}
		state.clients[key] = client
	}
	client.LastSeen = time.Now().Unix()
	client.ActiveRequests++
	state.mu.Unlock()
	listening := false
	started := func() {
		state.mu.Lock()
		defer state.mu.Unlock()
		if !listening {
			listening = true
			if client.Listeners == 0 {
				client.ConnectedSince = time.Now().Unix()
			}
			client.Listeners++
		}
	}
	finished := func() {
		state.mu.Lock()
		defer state.mu.Unlock()
		client.ActiveRequests--
		if listening {
			client.Listeners--
			if client.Listeners == 0 {
				client.ConnectedSince = 0
			}
		}
		client.LastSeen = time.Now().Unix()
	}
	return started, finished
}
