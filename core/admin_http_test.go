package core

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chanify/chanify/logic"
)

func TestAdminHTTPAccessAcrossAddresses(t *testing.T) {
	c := New()
	if err := c.Init(&logic.Options{DBUrl: "sqlite://?mode=memory", Endpoint: "https://node.example", AdminPassword: "synthetic-admin-password"}); err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	h := c.APIHandler()
	for _, base := range []string{"http://10.1.2.3", "http://172.20.1.2", "http://192.168.1.2", "http://node.internal", "http://[fd00::1234]", "https://node.internal"} {
		t.Run(base, func(t *testing.T) {
			request := func(method, path string, body any, cookie *http.Cookie, csrf string) *httptest.ResponseRecorder {
				data, err := json.Marshal(body)
				if err != nil {
					t.Fatal(err)
				}
				r := httptest.NewRequest(method, base+path, bytes.NewReader(data))
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Origin", base)
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
			w := request("GET", "/admin/api/config", nil, nil, "")
			if w.Code != 200 || !bytes.Contains(w.Body.Bytes(), []byte(`"allow_insecure_http":true`)) {
				t.Fatal(w.Code, w.Body.String())
			}
			w = request("POST", "/admin/api/login", map[string]string{"password_digest": adminDigest("synthetic-admin-password")}, nil, "")
			if w.Code != 200 {
				t.Fatal(w.Code, w.Body.String())
			}
			cookies := w.Result().Cookies()
			if len(cookies) != 1 {
				t.Fatal("missing cookie")
			}
			cookie := cookies[0]
			if cookie.Secure != (base == "https://node.internal") || !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode {
				t.Fatal("wrong cookie security", cookie)
			}
			var session struct {
				CSRF string `json:"csrf"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &session); err != nil || session.CSRF == "" {
				t.Fatal("missing CSRF", err)
			}
			if w := request("GET", "/admin/api/session", nil, cookie, ""); w.Code != 200 {
				t.Fatal(w.Code)
			}
			if w := request("POST", "/admin/api/logout", map[string]string{}, cookie, ""); w.Code != 403 {
				t.Fatal("missing CSRF accepted", w.Code)
			}
			w = request("POST", "/admin/api/logout", map[string]string{}, cookie, session.CSRF)
			if w.Code != 200 || w.Result().Cookies()[0].Secure != cookie.Secure {
				t.Fatal("logout failed", w.Code)
			}
			if w := request("GET", "/admin/api/session", nil, cookie, ""); w.Code != 401 {
				t.Fatal("session survived logout", w.Code)
			}
		})
	}
}

func TestAdminHTTPFallbackCanBeDisabled(t *testing.T) {
	c := New()
	if err := c.Init(&logic.Options{DBUrl: "sqlite://?mode=memory", Endpoint: "https://node.example", AdminPassword: "synthetic-admin-password", AdminDisableInsecureHTTP: true}); err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	h := c.adminHandler()
	w := adminRequest(h, "GET", "/admin/api/config", nil, nil, "")
	if w.Code != 200 || !bytes.Contains(w.Body.Bytes(), []byte(`"allow_insecure_http":false`)) {
		t.Fatal(w.Code, w.Body.String())
	}
	cookie, _ := adminLoginTest(t, h)
	if !cookie.Secure {
		t.Fatal("default HTTPS cookie weakened")
	}
	w = adminRequest(h, "GET", "/admin/password.js", nil, nil, "")
	if w.Code != 200 || !bytes.Contains(w.Body.Bytes(), []byte("ChanifyPassword")) {
		t.Fatal("digest asset unavailable", w.Code)
	}
}
