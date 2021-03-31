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
	req := httptest.NewRequest("GET", "/not-found/1234567890/abcdefghijklmnopqrstuvwxyz/ABCDEFGHIJKLMNOPQRSTUVWXYZ", nil)
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
	c.Init(&logic.Options{DBUrl: "sqlite://?mode=memory", Registerable: true})                                                                                                                      // nolint: errcheck
	c.logic.UpsertUser("ABOO6TSIXKSEVIJKXLDQSUXQRXUAOXGGYY", "BGaP1ekObDB0bRkmvxkvfFXCLSk46mO7rW8PikP8sWsA_97yij0s0U7ioA9dWEoz41TrUP8Z88XzQ_Tl8AOoJF4", false)                                      // nolint: errcheck
	c.logic.BindDevice("ABOO6TSIXKSEVIJKXLDQSUXQRXUAOXGGYY", "B3BC1B875EDA13986801B1004B4ABF5760C197F4", "BDuFNLkmxyK0-NN3H3oKzzOtISq1w17-JAibD7X4pljYl6IEaEglWkKD5Iw537h-DYxAooXkHtu6un078sm7IiQ") // nolint: errcheck
	handler := c.APIHandler()
	s := `{"nonce": 123,"device": "B3BC1B875EDA13986801B1004B4ABF5760C197F4","user": "ABOO6TSIXKSEVIJKXLDQSUXQRXUAOXGGYY","token": ""}`
	req := httptest.NewRequest("POST", "/rest/v1/push-token", strings.NewReader(s))
	req.Header.Set("CHUserSign", "MEUCIH9gSXOY2ow1VWZjfqgpnXTJSWTV86hChjgPpKQFMpBuAiEArM1KZ5x2POO_XHrvltt30rIf6oX-YTBefShhaosK2TY")
	req.Header.Set("CHDevSign", "MEUCIB7Hjnl2_k-IGHIjB7HDeo5T55Sa1Sp6junm8o4jzE6HAiEAgz3QcjuEt22P1j1gQTRGNHwIgotgKtHOl54Daqd6AtU")
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
	c.Init(&logic.Options{DBUrl: "sqlite://?mode=memory", Registerable: true}) // nolint: errcheck
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
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("Check update push token invalid user id")
	}

	c.logic.UpsertUser("ABOO6TSIXKSEVIJKXLDQSUXQRXUAOXGGYY", "BGaP1ekObDB0bRkmvxkvfFXCLSk46mO7rW8PikP8sWsA_97yij0s0U7ioA9dWEoz41TrUP8Z88XzQ_Tl8AOoJF4", true) // nolint: errcheck
	req = httptest.NewRequest("POST", "/rest/v1/push-token", strings.NewReader(`{
		"nonce": 123,
		"device": "abc",
		"user": "ABOO6TSIXKSEVIJKXLDQSUXQRXUAOXGGYY",
		"token": "tk-string"
	}`))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp = w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("Check update push token invalid user mode")
	}

	c.logic.UpsertUser("ABOO6TSIXKSEVIJKXLDQSUXQRXUAOXGGYY", "BGaP1ekObDB0bRkmvxkvfFXCLSk46mO7rW8PikP8sWsA_97yij0s0U7ioA9dWEoz41TrUP8Z88XzQ_Tl8AOoJF4", false) // nolint: errcheck
	s := `{"nonce": 123,"device": "abc","user": "ABOO6TSIXKSEVIJKXLDQSUXQRXUAOXGGYY","token": "tk-string"}`
	req = httptest.NewRequest("POST", "/rest/v1/push-token", strings.NewReader(s))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp = w.Result()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Error("Check update push token invalid user sign")
	}

	req = httptest.NewRequest("POST", "/rest/v1/push-token", strings.NewReader(s))
	req.Header.Set("CHUserSign", "MEUCIQDFZqli_bzaW9MsPY6vjcOuAhrIlOg9c7Fl3G8adA9RqgIgM7BPNA-DHRnWdHkXn61BIrQIArLv4BS76TzBhvgqs2g")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp = w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("Check update push token invalid device")
	}

	c.logic.BindDevice("ABOO6TSIXKSEVIJKXLDQSUXQRXUAOXGGYY", "B3BC1B875EDA13986801B1004B4ABF5760C197F4", "BDuFNLkmxyK0-NN3H3oKzzOtISq1w17-JAibD7X4pljYl6IEaEglWkKD5Iw537h-DYxAooXkHtu6un078sm7IiQ") // nolint: errcheck
	s = `{"nonce": 123,"device": "B3BC1B875EDA13986801B1004B4ABF5760C197F4","user": "ABOO6TSIXKSEVIJKXLDQSUXQRXUAOXGGYY","token": "tk-string"}`
	req = httptest.NewRequest("POST", "/rest/v1/push-token", strings.NewReader(s))
	req.Header.Set("CHUserSign", "MEQCIF9UolxBEzndeJHMTe3N9dmcYoYUI9gv9uqmtfo-fewpAiBo0hszyxlvQo4_jUpFrHu2QoRug-SJNj3JfwWQD3HIrA")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp = w.Result()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Error("Check update push token invalid device sign")
	}

	req = httptest.NewRequest("POST", "/rest/v1/push-token", strings.NewReader(s))
	req.Header.Set("CHUserSign", "MEQCIF9UolxBEzndeJHMTe3N9dmcYoYUI9gv9uqmtfo-fewpAiBo0hszyxlvQo4_jUpFrHu2QoRug-SJNj3JfwWQD3HIrA")
	req.Header.Set("CHDevSign", "MEYCIQCk5wwXVh1L8H_ZOqdF8PptPpl5q6selyI8kP7xAw2oXQIhAPvZv0oJHkHYkrtcWH1RZg4xV-5Q0V-Omszqx7W2WeQo")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp = w.Result()
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
