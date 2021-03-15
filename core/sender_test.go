package core

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSender(t *testing.T) {
	c := New()
	defer c.Close()
	handler := c.APIHandler()
	req := httptest.NewRequest("POST", "/rest/v1/sender", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("Check health failed")
	}
}

func TestSenderPost(t *testing.T) {
	c := New()
	defer c.Close()
	handler := c.APIHandler()
	req := httptest.NewRequest("POST", "/rest/v1/sender", bytes.NewReader([]byte("Hello")))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("Check health failed")
	}
}

func TestSenderPostForm(t *testing.T) {
	c := New()
	defer c.Close()
	handler := c.APIHandler()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	partText, _ := writer.CreateFormField("text")
	partText.Write([]byte("hello"))
	partToken, _ := writer.CreateFormField("token")
	partToken.Write([]byte("token"))
	writer.Close()

	req := httptest.NewRequest("POST", "/rest/v1/sender", body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("Check health failed")
	}
}
