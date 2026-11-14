package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/chanify/chanify/logic"
	"github.com/chanify/chanify/model"
)

type mediaListResponse struct {
	Items    []logic.MediaFile `json:"items"`
	Total    int               `json:"total"`
	Count    int               `json:"count"`
	Page     int               `json:"page"`
	Pages    int               `json:"pages"`
	Bytes    int64             `json:"bytes"`
	Snapshot string            `json:"snapshot"`
}

func TestAdminMediaAuthenticationAndDisabledStorage(t *testing.T) {
	c, _, _ := mcpFixture(t, false, false)
	c.admin = newAdmin(testAdminHash(t), "http://localhost")
	h := c.APIHandler()
	cookie, csrf := adminLoginTest(t, h)
	for _, path := range []string{"/admin/api/media", "/admin/api/media/content?id=files/" + strings.Repeat("a", 40)} {
		if w := adminRequest(h, "GET", path, nil, nil, ""); w.Code != 401 {
			t.Fatal(w.Code)
		}
		if w := adminRequest(h, "GET", path, nil, cookie, ""); w.Code != 404 {
			t.Fatal(w.Code)
		}
	}
	if w := adminRequest(h, "POST", "/admin/api/media", map[string]interface{}{"action": "clear", "snapshot": "x"}, cookie, ""); w.Code != 403 {
		t.Fatal(w.Code)
	}
	if w := adminRequest(h, "POST", "/admin/api/media", map[string]interface{}{"action": "clear", "snapshot": "x"}, cookie, csrf); w.Code != 404 {
		t.Fatal(w.Code)
	}
	w := adminRequest(h, "GET", "/admin/api/state", nil, cookie, "")
	var state struct {
		Capabilities struct {
			Media bool `json:"media"`
		} `json:"capabilities"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &state); err != nil || state.Capabilities.Media {
		t.Fatal(w.Body.String(), err)
	}
}
func TestAdminMediaSearchPagingDownloadAndClear(t *testing.T) {
	c, _, _ := mcpFixture(t, true, false)
	c.admin = newAdmin(testAdminHash(t), "http://localhost")
	h := c.APIHandler()
	cookie, csrf := adminLoginTest(t, h)
	for i := 0; i < 51; i++ {
		if _, err := c.logic.SaveFile("files", []byte(fmt.Sprintf("synthetic %03d", i)), fmt.Sprintf("report-%03d.txt", i)); err != nil {
			t.Fatal(err)
		}
	}
	list := func(query string) mediaListResponse {
		t.Helper()
		w := adminRequest(h, "GET", "/admin/api/media"+query, nil, cookie, "")
		if w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
		var response mediaListResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		return response
	}
	all := list("?sort=name")
	if all.Count != 51 || all.Total != 51 || len(all.Items) != 50 || all.Pages != 2 {
		t.Fatal(all)
	}
	last := list("?sort=name&page=999999999")
	if last.Page != 2 || len(last.Items) != 1 {
		t.Fatal(last)
	}
	found := list("?q=REPORT-000&kind=text")
	if found.Total != 1 {
		t.Fatal(found)
	}
	item := found.Items[0]
	if _, err := c.logic.SaveFile("files", []byte("synthetic 000"), "alias.txt"); err != nil {
		t.Fatal(err)
	}
	if list("?q=alias.txt").Total != 1 {
		t.Fatal("alias not searchable")
	}
	path := "/admin/api/media/content?id=" + url.QueryEscape(item.ID)
	w := adminRequest(h, "GET", path, nil, cookie, "")
	if w.Code != 200 || w.Body.String() != "synthetic 000" || !strings.HasPrefix(w.Header().Get("Content-Disposition"), "attachment;") {
		t.Fatal(w.Code, w.Header(), w.Body.String())
	}
	req := httptest.NewRequest("GET", path, nil)
	req.AddCookie(cookie)
	req.Header.Set("Range", "bytes=0-3")
	w = httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusPartialContent || w.Body.String() != "synt" {
		t.Fatal(w.Code, w.Body.String())
	}
	w = adminRequest(h, "POST", "/admin/api/media", map[string]interface{}{"action": "clear", "snapshot": all.Snapshot}, cookie, csrf)
	if w.Code != 409 {
		t.Fatal(w.Code, w.Body.String())
	}
	for _, body := range []interface{}{map[string]interface{}{"action": "clear"}, map[string]interface{}{"action": "delete", "ids": []string{item.ID, "files/../outside"}}, map[string]interface{}{"action": "delete", "ids": []string{}}, map[string]interface{}{"action": "clear", "snapshot": "x", "ids": []string{item.ID}}} {
		if w := adminRequest(h, "POST", "/admin/api/media", body, cookie, csrf); w.Code != 400 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	if list("").Count != 51 {
		t.Fatal("invalid request deleted files")
	}
	w = adminRequest(h, "POST", "/admin/api/media", map[string]interface{}{"action": "delete", "ids": []string{item.ID, "files/" + strings.Repeat("0", 40)}}, cookie, csrf)
	var result logic.MediaDeleteResult
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || result.Deleted != 1 || len(result.Failed) != 1 {
		t.Fatal(w.Body.String(), err)
	}
	if w := adminRequest(h, "GET", path, nil, cookie, ""); w.Code != 404 {
		t.Fatal(w.Code)
	}
	snapshot := list("").Snapshot
	w = adminRequest(h, "POST", "/admin/api/media", map[string]interface{}{"action": "clear", "snapshot": snapshot}, cookie, csrf)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if list("").Count != 0 {
		t.Fatal("clear left files")
	}
}
func TestAdminMediaSafePreviewAndTextLimit(t *testing.T) {
	c, _, _ := mcpFixture(t, true, false)
	c.admin = newAdmin(testAdminHash(t), "http://localhost")
	h := c.APIHandler()
	cookie, _ := adminLoginTest(t, h)
	for _, tc := range []struct {
		name                  string
		data                  []byte
		wantType, disposition string
		limit                 int
	}{
		{"report.txt", bytes.Repeat([]byte("a"), 300<<10), "text/plain; charset=utf-8", "inline", 256 << 10},
		{"image.png", []byte("<!doctype html><script>alert(1)</script>"), "application/octet-stream", "attachment", 0},
		{"vector.svg", []byte("<?xml version=\"1.0\"?><svg xmlns=\"http://www.w3.org/2000/svg\"></svg>"), "application/octet-stream", "attachment", 0},
	} {
		path, err := c.logic.SaveFile("files", tc.data, tc.name)
		if err != nil {
			t.Fatal(err)
		}
		w := adminRequest(h, "GET", "/admin/api/media/content?preview=1&id="+url.QueryEscape(strings.TrimPrefix(path, "/files/")), nil, cookie, "")
		if w.Code != 200 || w.Header().Get("Content-Type") != tc.wantType || !strings.HasPrefix(w.Header().Get("Content-Disposition"), tc.disposition+";") {
			t.Fatal(tc.name, w.Code, w.Header())
		}
		if tc.limit > 0 && w.Body.Len() != tc.limit {
			t.Fatal(tc.name, w.Body.Len())
		}
	}
	if w := adminRequest(h, "GET", "/admin/api/media/content?id=../config.yml", nil, cookie, ""); w.Code != 404 {
		t.Fatal(w.Code)
	}
}

func TestMediaLongTextUploadMetadata(t *testing.T) {
	c, _, _ := mcpFixture(t, true, false)
	// Multi-byte titles and short bodies must not panic while generating attachment previews.
	for _, tc := range []struct{ title, text string }{
		{strings.Repeat("报", 40), strings.Repeat("body", 600)},
		{strings.Repeat("题", 1000), "short"},
	} {
		if _, err := c.makeTextContent(new(model.Message), tc.text, tc.title, "", "", nil); err != nil {
			t.Fatal(err)
		}
	}
	inventory, err := c.logic.MediaInventory()
	if err != nil || inventory.Count != 2 {
		t.Fatal(inventory, err)
	}
	for _, item := range inventory.Items {
		if item.Legacy || item.Kind != "text" || len(item.Names) != 1 || !strings.HasSuffix(item.Names[0], ".txt") {
			t.Fatal(item)
		}
	}
}
