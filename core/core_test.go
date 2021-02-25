package core

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCore(t *testing.T) {
	c := New()
	if c == nil {
		t.Error("Create core failed!")
	}
	defer c.Close()
	handler := c.APIHandler()
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Error("Check health failed")
	}
}
