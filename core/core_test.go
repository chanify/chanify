package core

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chanify/chanify/logic"
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
	req := httptest.NewRequest("GET", "/not-found", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Error("Check not found failed")
	}
}

func TestUpdatePushToken(t *testing.T) {
	c := New()
	defer c.Close()
	c.Init(&logic.Options{DBUrl: "sqlite://?mode=memory"})                                                                                                                                          // nolint: errcheck
	c.logic.UpsertUser("ABOO6TSIXKSEVIJKXLDQSUXQRXUAOXGGYY", "BGaP1ekObDB0bRkmvxkvfFXCLSk46mO7rW8PikP8sWsA_97yij0s0U7ioA9dWEoz41TrUP8Z88XzQ_Tl8AOoJF4", true)                                       // nolint: errcheck
	c.logic.BindDevice("ABOO6TSIXKSEVIJKXLDQSUXQRXUAOXGGYY", "B3BC1B875EDA13986801B1004B4ABF5760C197F4", "BDuFNLkmxyK0-NN3H3oKzzOtISq1w17-JAibD7X4pljYl6IEaEglWkKD5Iw537h-DYxAooXkHtu6un078sm7IiQ") // nolint: errcheck
	handler := c.APIHandler()
	req := httptest.NewRequest("POST", "/rest/v1/push-token", strings.NewReader(`{
		"nonce": 123,
		"device": "B3BC1B875EDA13986801B1004B4ABF5760C197F4",
		"user": "ABOO6TSIXKSEVIJKXLDQSUXQRXUAOXGGYY",
		"token": ""
	}`))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Error("Check update push token failed")
	}
}

func TestUpdatePushTokenFailed(t *testing.T) {
	c := New()
	defer c.Close()
	c.Init(&logic.Options{DBUrl: "sqlite://?mode=memory"}) // nolint: errcheck
	handler := c.APIHandler()
	req := httptest.NewRequest("POST", "/rest/v1/push-token", strings.NewReader(`{
		"nonce": 123,
		"device": "abc",
		"user": "xyz",
		"token": "tk-string"
	}`))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusConflict {
		t.Error("Check update push token failed")
	}
}

func TestUpdatePushTokenInvalid(t *testing.T) {
	c := New()
	defer c.Close()
	handler := c.APIHandler()
	req := httptest.NewRequest("POST", "/rest/v1/push-token", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("Check update push token request failed")
	}
}
