package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/chanify/chanify/logic"
)

func TestAdminDecodeRejectsMalformedBodies(t *testing.T) {
	var body interface{}
	if err := adminDecode(httptest.NewRecorder(), httptest.NewRequest("POST", "/", strings.NewReader(`{}`)), &body); err == nil {
		t.Fatal("non-JSON content type accepted")
	}
	for _, body := range []string{`{"value":1} {"value":2}`, `{"unknown":1}`, `{`, strings.Repeat(" ", 4097) + `{}`} {
		r := httptest.NewRequest("POST", "/", strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		var value struct {
			Value int `json:"value"`
		}
		if err := adminDecode(httptest.NewRecorder(), r, &value); err == nil {
			t.Fatal("malformed body accepted")
		}
	}
}

func TestAdminLoginResourceLimits(t *testing.T) {
	for _, mode := range []string{"digest", "slots", "attempts", "sessions", "expired"} {
		t.Run(mode, func(t *testing.T) {
			c := New()
			c.admin = newAdmin(testAdminHash(t), "http://localhost")
			digest := adminDigest("synthetic-admin-password")
			switch mode {
			case "digest":
				digest = strings.Repeat("z", 64)
			case "slots":
				c.admin.loginSlots <- struct{}{}
				c.admin.loginSlots <- struct{}{}
			case "attempts":
				for i := 0; i < 1024; i++ {
					c.admin.attempts[fmt.Sprint(i)] = adminAttempt{Until: time.Now().Add(time.Hour)}
				}
			case "sessions":
				for i := 0; i < 256; i++ {
					c.admin.sessions[fmt.Sprint(i)] = adminSession{Expires: time.Now().Add(time.Hour)}
				}
			case "expired":
				c.admin.attempts["old"] = adminAttempt{Until: time.Now().Add(-time.Hour)}
				c.admin.sessions["old"] = adminSession{Expires: time.Now().Add(-time.Hour)}
			}
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { r.RemoteAddr = "bare-address"; c.adminLogin(w, r) })
			w := adminRequest(h, "POST", "/", map[string]string{"password_digest": digest}, &http.Cookie{Name: adminCookie, Value: "old-session"}, "")
			want := 429
			if mode == "digest" {
				want = 400
			}
			if mode == "expired" {
				want = 200
			}
			if w.Code != want {
				t.Fatal(w.Code, w.Body.String())
			}
			if mode == "expired" {
				if _, ok := c.admin.sessions["old"]; ok {
					t.Fatal("expired session retained")
				}
				if _, ok := c.admin.attempts["old"]; ok {
					t.Fatal("expired attempt retained")
				}
			}
		})
	}
}

func TestAdminTokenFailuresAndSnapshotClients(t *testing.T) {
	c, _, _ := mcpFixture(t, false, false)
	c.admin = newAdmin(testAdminHash(t), "http://localhost")
	h := http.HandlerFunc(c.adminTokens)
	for _, body := range []map[string]interface{}{
		{"unknown": true}, {"action": "create", "kind": "mcp", "user": "missing", "name": "name", "expires": time.Now().Add(time.Hour).Unix()},
		{"action": "delete"}, {"action": "rename", "kind": "unknown"},
	} {
		if w := adminRequest(h, "POST", "/", body, nil, ""); w.Code != 400 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	c.admin.clients["one"] = &adminClient{ID: "one", LastSeen: 1}
	c.admin.clients["two"] = &adminClient{ID: "two", LastSeen: 2}
	w := httptest.NewRecorder()
	c.adminSnapshot(w, httptest.NewRequest("GET", "/", nil))
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"id":"two"`) {
		t.Fatal(w.Code, w.Body.String())
	}
}

func TestAdminActivityCapacity(t *testing.T) {
	for _, busy := range []bool{false, true} {
		c := New()
		c.admin = newAdmin("", "http://localhost")
		for i := 0; i < 2048; i++ {
			active := 0
			if busy {
				active = 1
			}
			c.admin.clients[fmt.Sprint(i)] = &adminClient{LastSeen: int64(i), ActiveRequests: active}
		}
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = "bare-address"
		r.Header.Set("User-Agent", strings.Repeat("x", 300))
		start, finish := c.beginAdminActivity(r, "mcp", "id", "user")
		start()
		finish()
		if len(c.admin.clients) != 2048 {
			t.Fatal("activity limit exceeded")
		}
		if !busy {
			if _, ok := c.admin.clients["0"]; ok {
				t.Fatal("oldest client retained")
			}
		}
	}
}

func TestMediaFallbackValues(t *testing.T) {
	m := logic.MediaFile{ID: "files/id", Modified: 123}
	if mediaDisplayName(m) != m.ID || mediaTime(m) != 123 {
		t.Fatal("fallback changed")
	}
}

func TestAdminMediaSorting(t *testing.T) {
	c, _, _ := mcpFixture(t, true, false)
	for _, tc := range []struct{ body, name string }{{"small", "z.txt"}, {"some longer content", "a.txt"}, {"equal", "b.txt"}} {
		if _, err := c.logic.SaveFile("files", []byte(tc.body), tc.name); err != nil {
			t.Fatal(err)
		}
	}
	for _, sortBy := range []string{"size", "name", "oldest", "newest"} {
		r := httptest.NewRequest("GET", "/admin/api/media?sort="+sortBy+"&page=999", nil)
		w := httptest.NewRecorder()
		c.adminMediaList(w, r)
		if w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
		var result struct {
			Items       []logic.MediaFile `json:"items"`
			Total, Page int
		}
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if result.Total != 3 || result.Page != 1 {
			t.Fatal(result)
		}
		if sortBy == "size" && result.Items[0].Size != 19 {
			t.Fatal(result.Items)
		}
		if sortBy == "name" && mediaDisplayName(result.Items[0]) != "a.txt" {
			t.Fatal(result.Items)
		}
	}
}

func TestAdminRandomAndSessionFailures(t *testing.T) {
	if _, err := adminRandomFrom(strings.NewReader("short")); err == nil {
		t.Fatal("incomplete entropy accepted")
	}
	for _, failAt := range []int{0, 1} {
		c := New()
		c.admin = newAdmin(testAdminHash(t), "http://localhost")
		calls := 0
		c.admin.random = func() (string, error) {
			if calls == failAt {
				return "", io.ErrUnexpectedEOF
			}
			calls++
			return "synthetic", nil
		}
		w := adminRequest(http.HandlerFunc(c.adminLogin), "POST", "/", map[string]string{"password_digest": adminDigest("synthetic-admin-password")}, nil, "")
		if w.Code != 500 || len(c.admin.sessions) != 0 || len(w.Result().Cookies()) != 0 {
			t.Fatal(w.Code, "failed login created a session")
		}
	}
}

func TestAdminMissingAssetsAndSessionLookup(t *testing.T) {
	c := New()
	c.admin = newAdmin("", "http://localhost")
	c.admin.assets = fstest.MapFS{}
	w := adminRequest(c.adminHandler(), "GET", "/admin", nil, nil, "")
	if w.Code != 404 {
		t.Fatal(w.Code)
	}
	c.admin.sessions[adminDigest("synthetic-session")] = adminSession{CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	w = adminRequest(c.adminHandler(), "GET", "/admin/api/session", nil, &http.Cookie{Name: adminCookie, Value: "synthetic-session"}, "")
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"csrf":"csrf"`) {
		t.Fatal(w.Code, w.Body.String())
	}
}

func TestAdminLegacyMediaOrderingAndPreview(t *testing.T) {
	root := t.TempDir()
	c := New()
	if err := c.Init(&logic.Options{DBUrl: "sqlite://?mode=memory", FilePath: root}); err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	for i := 0; i < 3; i++ {
		path := filepath.Join(root, "files", strings.Repeat(string(rune('a'+i)), 40))
		if err := os.WriteFile(path, []byte("data"), 0600); err != nil {
			t.Fatal(err)
		}
		stamp := time.Unix(1000+int64(i), 0)
		if err := os.Chtimes(path, stamp, stamp); err != nil {
			t.Fatal(err)
		}
	}
	for _, sortBy := range []string{"oldest", "newest"} {
		w := httptest.NewRecorder()
		c.adminMediaList(w, httptest.NewRequest("GET", "/?sort="+sortBy, nil))
		var result struct {
			Items []logic.MediaFile `json:"items"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		first := "a"
		if sortBy == "newest" {
			first = "c"
		}
		if len(result.Items) != 3 || !strings.HasSuffix(result.Items[0].ID, strings.Repeat(first, 40)) {
			t.Fatal(result.Items)
		}
	}
	w := httptest.NewRecorder()
	c.adminMediaList(w, httptest.NewRequest("GET", "/?kind=image", nil))
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"total":0`) {
		t.Fatal(w.Code, w.Body.String())
	}
	var data bytes.Buffer
	if err := png.Encode(&data, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil {
		t.Fatal(err)
	}
	url, err := c.logic.SaveFile("images", data.Bytes(), "image.png")
	if err != nil {
		t.Fatal(err)
	}
	w = httptest.NewRecorder()
	c.adminMediaContent(w, httptest.NewRequest("GET", "/?preview=1&id="+strings.TrimPrefix(url, "/files/"), nil))
	if w.Code != 200 || w.Header().Get("Content-Type") != "image/png" || !strings.HasPrefix(w.Header().Get("Content-Disposition"), "inline") {
		t.Fatal(w.Code, w.Header())
	}
	f, err := os.CreateTemp(t.TempDir(), "closed")
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	w = httptest.NewRecorder()
	serveAdminMedia(w, httptest.NewRequest("GET", "/", nil), f, logic.MediaFile{})
	if w.Code != 404 {
		t.Fatal(w.Code)
	}
}

func TestAdminRequiresPersistentStorage(t *testing.T) {
	c := New()
	if err := c.Init(&logic.Options{Secret: "synthetic", AdminPassword: "synthetic-admin-password"}); err == nil {
		c.Close()
		t.Fatal("stateless administration accepted")
	}
}
