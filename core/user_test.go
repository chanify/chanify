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
	c.Init(&logic.Options{DBUrl: "sqlite://?mode=memory", Registerable: true}) // nolint: errcheck
	handler := c.APIHandler()
	s := `{"user": {"uid": "ABOO6TSIXKSEVIJKXLDQSUXQRXUAOXGGYY","key": "BGaP1ekObDB0bRkmvxkvfFXCLSk46mO7rW8PikP8sWsA_97yij0s0U7ioA9dWEoz41TrUP8Z88XzQ_Tl8AOoJF4"}}`
	req := httptest.NewRequest("POST", "/rest/v1/bind-user", strings.NewReader(s))
	req.Header.Set("CHUserSign", "MEUCIQDD93w25DdEJCIkIZU5GioFFAvTBILvuq3l-YBbapMOpQIgKJaszx-jwcWjhADsD2XlWTLtLlBPSTUch9LoNP0pS9Y")
	req.Header.Set("CHDevSign", "MEQCIEqo-nBRlEempp1U43xfGMYzRbWEvnJXcROAZP2dpuWtAiBIicKZgDYNpc6y7Ihov9w21EK8CTPztNx0c_4pmz5ehA")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Error("Bind user failed")
	}

	req = httptest.NewRequest("POST", "/rest/v1/bind-user", strings.NewReader(s))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp = w.Result()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Error("Check bind user sign failed")
	}

	s = `{"device": {"uuid": "B3BC1B875EDA13986801B1004B4ABF5760C197F4","key": "BDuFNLkmxyK0-NN3H3oKzzOtISq1w17-JAibD7X4pljYl6IEaEglWkKD5Iw537h-DYxAooXkHtu6un078sm7IiQ","push-token": "aGVsbG8"},"user": {"uid": "ABOO6TSIXKSEVIJKXLDQSUXQRXUAOXGGYY","key": "BGaP1ekObDB0bRkmvxkvfFXCLSk46mO7rW8PikP8sWsA_97yij0s0U7ioA9dWEoz41TrUP8Z88XzQ_Tl8AOoJF4"}}`
	req = httptest.NewRequest("POST", "/rest/v1/bind-user", strings.NewReader(s))
	req.Header.Set("CHUserSign", "MEYCIQD-4jUyN0NuBJ_U9rjmPNNf36QWy-l05tZazyO1k23sHAIhAPmgikDQGovVb1GZll4LkfaavJ74eIN6UuTEbvgNowLj")
	req.Header.Set("CHDevSign", "MEQCIGaFG_etoxnari4rSz-ZHvNTLd9hlBk_pb2N4kuqE2HgAiBDlVxuI22K7B-CpYoLIJWXLNZfJeoigHyUFalcn5j60A")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Error("Check bind user failed")
	}

	req = httptest.NewRequest("POST", "/rest/v1/bind-user", strings.NewReader(s))
	req.Header.Set("CHUserSign", "MEYCIQD-4jUyN0NuBJ_U9rjmPNNf36QWy-l05tZazyO1k23sHAIhAPmgikDQGovVb1GZll4LkfaavJ74eIN6UuTEbvgNowLj")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp = w.Result()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Error("Check bind user device sign failed")
	}
}

func TestBindUserLimited(t *testing.T) {
	c := New()
	defer c.Close()
	c.Init(&logic.Options{ // nolint: errcheck
		DBUrl:        "sqlite://?mode=memory",
		Registerable: false,
		RegUsers:     []string{"ABOO6TSIXKSEVIJKXLDQSUXQRXUAOXGGYY"},
	})
	handler := c.APIHandler()
	s := `{"user": {"uid": "ABOO6TSIXKSEVIJKXLDQSUXQRXUAOXGGYY","key": "BGaP1ekObDB0bRkmvxkvfFXCLSk46mO7rW8PikP8sWsA_97yij0s0U7ioA9dWEoz41TrUP8Z88XzQ_Tl8AOoJF4"}}`
	req := httptest.NewRequest("POST", "/rest/v1/bind-user", strings.NewReader(s))
	req.Header.Set("CHUserSign", "MEUCIQDD93w25DdEJCIkIZU5GioFFAvTBILvuq3l-YBbapMOpQIgKJaszx-jwcWjhADsD2XlWTLtLlBPSTUch9LoNP0pS9Y")
	req.Header.Set("CHDevSign", "MEQCIEqo-nBRlEempp1U43xfGMYzRbWEvnJXcROAZP2dpuWtAiBIicKZgDYNpc6y7Ihov9w21EK8CTPztNx0c_4pmz5ehA")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Error("Bind limited user failed")
	}
}

func TestBindUserLimitedFailed(t *testing.T) {
	c := New()
	defer c.Close()
	c.Init(&logic.Options{DBUrl: "sqlite://?mode=memory", Registerable: false}) // nolint: errcheck
	handler := c.APIHandler()
	s := `{"user": {"uid": "ABOO6TSIXKSEVIJKXLDQSUXQRXUAOXGGYY","key": "BGaP1ekObDB0bRkmvxkvfFXCLSk46mO7rW8PikP8sWsA_97yij0s0U7ioA9dWEoz41TrUP8Z88XzQ_Tl8AOoJF4"}}`
	req := httptest.NewRequest("POST", "/rest/v1/bind-user", strings.NewReader(s))
	req.Header.Set("CHUserSign", "MEUCIQDD93w25DdEJCIkIZU5GioFFAvTBILvuq3l-YBbapMOpQIgKJaszx-jwcWjhADsD2XlWTLtLlBPSTUch9LoNP0pS9Y")
	req.Header.Set("CHDevSign", "MEQCIEqo-nBRlEempp1U43xfGMYzRbWEvnJXcROAZP2dpuWtAiBIicKZgDYNpc6y7Ihov9w21EK8CTPztNx0c_4pmz5ehA")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusNotAcceptable {
		t.Error("Check bind user limit failed")
	}
}

func TestBindUserFailed(t *testing.T) {
	c := New()
	defer c.Close()
	c.Init(&logic.Options{DBUrl: "sqlite://?mode=memory", Registerable: true}) // nolint: errcheck
	handler := c.APIHandler()
	req := httptest.NewRequest("POST", "/rest/v1/bind-user", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("Check bind user failed")
	}

	s := `{"user": {"uid": "ABOO6TSIXKSEVIJKXLDQSUXQRXUAOXGGYX","key": "BGaP1ekObDB0bRkmvxkvfFXCLSk46mO7rW8PikP8sWsA_97yij0s0U7ioA9dWEoz41TrUP8Z88XzQ_Tl8AOoJF4"}}`
	req = httptest.NewRequest("POST", "/rest/v1/bind-user", strings.NewReader(s))
	req.Header.Set("CHUserSign", "MEYCIQDxfsNx3HyxbEBDd2oFzerNUIuNziQwmM-4gN12k5pTBAIhAKijSV4OEYabQplSHL-BLsMBhiBsVhDryRLq8wvB90On")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp = w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("Check bind user invalid user key")
	}

	s = `{"device": {"uuid": "B3BC1B875EDA13986801B1004B4ABF5760C197F5","key": "BDuFNLkmxyK0-NN3H3oKzzOtISq1w17-JAibD7X4pljYl6IEaEglWkKD5Iw537h-DYxAooXkHtu6un078sm7IiQ"},"user": {"uid": "ABOO6TSIXKSEVIJKXLDQSUXQRXUAOXGGYY","key": "BGaP1ekObDB0bRkmvxkvfFXCLSk46mO7rW8PikP8sWsA_97yij0s0U7ioA9dWEoz41TrUP8Z88XzQ_Tl8AOoJF4"}}`
	req = httptest.NewRequest("POST", "/rest/v1/bind-user", strings.NewReader(s))
	req.Header.Set("CHUserSign", "MEYCIQCIlIFubhPz7sI1cSFg79eZwT74MfQw4Jy3F7RF5R8JYwIhAJBI2gquLtqr50zrAFPurGVBrb1x7hpc6zEguAmWbkbj")
	req.Header.Set("CHDevSign", "MEYCIQDgy8kti33PYuuG2mbTWiSWFFUmPyZEBUtDp-l375oT5QIhAIc1nzxI22prTdVFX8A5M5HW7Ggoq9ZzDQ2Aqa7BRpNA")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp = w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("Check bind user invalid device key")
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

func TestUnbindUserInvalidSign(t *testing.T) {
	c := New()
	defer c.Close()
	c.Init(&logic.Options{DBUrl: "sqlite://?mode=memory", Registerable: true})                                                                                 // nolint: errcheck
	c.logic.UpsertUser("ABOO6TSIXKSEVIJKXLDQSUXQRXUAOXGGYY", "BGaP1ekObDB0bRkmvxkvfFXCLSk46mO7rW8PikP8sWsA_97yij0s0U7ioA9dWEoz41TrUP8Z88XzQ_Tl8AOoJF4", false) // nolint: errcheck
	handler := c.APIHandler()
	req := httptest.NewRequest("POST", "/rest/v1/unbind-user", strings.NewReader(`{
		"nonce": 123,
		"device": "abc",
		"user": "ABOO6TSIXKSEVIJKXLDQSUXQRXUAOXGGYY"
	}`))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusUnauthorized {
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
