package core

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetInfo(t *testing.T) {
	c := New()
	defer c.Close()
	c.SetSecret("secret")
	if len(c.info.NodeId) <= 0 {
		t.Error("Init NodeId failed")
	}
	c.SetVersion("0.1.2")
	if c.info.Version != "0.1.2" {
		t.Error("Set version failed")
	}
	c.SetEndpoint("http://127.0.0.1")
	if c.info.Endpoint != "http://127.0.0.1" || len(c.info.qrCode) <= 0 {
		t.Error("Set endpoint failed")
	}
	c.SetName("name-test")
	if c.info.Name != "name-test" {
		t.Error("Set name failed")
	}
	c.InitFeatures()
	if len(c.info.Features) <= 0 {
		t.Error("Set features failed")
	}
	handler := c.APIHandler()
	req := httptest.NewRequest("GET", "/rest/v1/qrcode", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Error("Get qrcode failed")
	}
	resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	if !bytes.Equal(data, c.info.qrCode) {
		t.Error("Invalid qrcode")
	}
	// if len(c.info.qrCode) > 0 {
	// 	t.Error("Check endpoint failed:", len(c.info.qrCode))
	// }
}

func TestInfo(t *testing.T) {
	c := New()
	defer c.Close()
	handler := c.APIHandler()
	req := httptest.NewRequest("GET", "/rest/v1/info", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Error("Check health failed")
	}
}
