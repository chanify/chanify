package core

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chanify/chanify/logic"
	"github.com/gin-gonic/gin"
)

func TestHealth(t *testing.T) {
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

func TestHome(t *testing.T) {
	c := New()
	defer c.Close()
	if err := c.Init(&logic.Options{Secret: "123"}); err != nil {
		t.Fatal("init core failed")
	}
	handler := c.APIHandler()
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Error("Check health failed:", resp.StatusCode)
	}
}

func TestInfo(t *testing.T) {
	c := New()
	defer c.Close()
	if err := c.Init(&logic.Options{Secret: "123"}); err != nil {
		t.Fatal("init core failed")
	}
	handler := c.APIHandler()
	req := httptest.NewRequest("GET", "/rest/v1/info", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Error("Check info failed:", resp.StatusCode)
	}
}

func TestNotFound(t *testing.T) {
	c := New()
	defer c.Close()
	handler := c.APIHandler()
	req := httptest.NewRequest("GET", "/not-found/1234567890/abcdefghijklmnopqrstuvwxyz/ABCDEFGHIJKLMNOPQRSTUVWXYZ", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Error("Check not found failed")
	}
}

func TestClientIP(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("GET", "", nil)
	ctx.Request.Header.Set("X-Real-IP", "10.10.10.100,127.0.0.200")
	if fixClientIP(ctx) != "10.10.10.100" {
		t.Error("Get client ip failed")
	}
	ctx.Request.Header.Set("X-Forwarded-For", "10.10.10.200")
	if fixClientIP(ctx) != "10.10.10.200" {
		t.Error("Check client ip order failed")
	}
	ctx.Request.Header.Set("X-Forwarded-For", "10.10.10.300")
	if fixClientIP(ctx) != "10.10.10.100" {
		t.Error("Check invalid client ip format failed")
	}
}
