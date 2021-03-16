package core

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUser(t *testing.T) {
	c := New("", "", "")
	defer c.Close()
	handler := c.APIHandler()
	req := httptest.NewRequest("POST", "/rest/v1/bind-user", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("Check health failed")
	}
}
