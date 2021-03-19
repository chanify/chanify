package core

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chanify/chanify/logic"
)

func TestBindUser(t *testing.T) {
	c := New()
	defer c.Close()
	c.Init(&logic.Options{DBUrl: "sqlite://?mode=memory"}) // nolint: errcheck
	handler := c.APIHandler()
	req := httptest.NewRequest("POST", "/rest/v1/bind-user", strings.NewReader(`{
		"user": {
			"uid": "ABOO6TSIXKSEVIJKXLDQSUXQRXUAOXGGYY",
			"key": "BGaP1ekObDB0bRkmvxkvfFXCLSk46mO7rW8PikP8sWsA_97yij0s0U7ioA9dWEoz41TrUP8Z88XzQ_Tl8AOoJF4"
		}
	}`))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Error("Bind user failed")
	}

	req = httptest.NewRequest("POST", "/rest/v1/bind-user", strings.NewReader(`{
		"device": {
			"uuid": "B3BC1B875EDA13986801B1004B4ABF5760C197F4",
			"key": "BDuFNLkmxyK0-NN3H3oKzzOtISq1w17-JAibD7X4pljYl6IEaEglWkKD5Iw537h-DYxAooXkHtu6un078sm7IiQ"
		},
		"user": {
			"uid": "ABOO6TSIXKSEVIJKXLDQSUXQRXUAOXGGYY",
			"key": "BGaP1ekObDB0bRkmvxkvfFXCLSk46mO7rW8PikP8sWsA_97yij0s0U7ioA9dWEoz41TrUP8Z88XzQ_Tl8AOoJF4"
		}
	}`))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Error("Check bind user failed")
	}
}

func TestBindUserFailed(t *testing.T) {
	c := New()
	defer c.Close()
	c.Init(&logic.Options{DBUrl: "sqlite://?mode=memory"}) // nolint: errcheck
	handler := c.APIHandler()
	req := httptest.NewRequest("POST", "/rest/v1/bind-user", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("Check bind user failed")
	}

	req = httptest.NewRequest("POST", "/rest/v1/bind-user", strings.NewReader(`{
		"user": {
			"uid": "ABOO6TSIXKSEVIJKXLDQSUXQRXUAOXGGYY",
			"key": "******"
		}
	}`))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp = w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("Check bind user invalid key failed")
	}

	req = httptest.NewRequest("POST", "/rest/v1/bind-user", strings.NewReader(`{
		"device": {
			"uuid": "B3BC1B875EDA13986801B1004B4ABF5760C197F4",
			"key": "*****"
		},
		"user": {
			"uid": "ABOO6TSIXKSEVIJKXLDQSUXQRXUAOXGGYY",
			"key": "BGaP1ekObDB0bRkmvxkvfFXCLSk46mO7rW8PikP8sWsA_97yij0s0U7ioA9dWEoz41TrUP8Z88XzQ_Tl8AOoJF4"
		}
	}`))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp = w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("Check bind user failed")
	}
}

func TestUnbindUser(t *testing.T) {
	c := New()
	defer c.Close()
	c.Init(&logic.Options{DBUrl: "sqlite://?mode=memory"}) // nolint: errcheck
	handler := c.APIHandler()
	req := httptest.NewRequest("POST", "/rest/v1/unbind-user", strings.NewReader(`{
		"nonce": 123,
		"device": "abc",
		"user": "xyz"
	}`))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Error("Ubind user failed")
	}
}

func TestUnbindUserFailed(t *testing.T) {
	c := New()
	defer c.Close()
	handler := c.APIHandler()
	req := httptest.NewRequest("POST", "/rest/v1/unbind-user", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("Check unbind user failed")
	}
}
