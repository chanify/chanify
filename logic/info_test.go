package logic

import "testing"

func TestInfo(t *testing.T) {
	l, _ := NewLogic(&Options{
		Name:     "name",
		Version:  "1.2.3",
		Endpoint: "http://127.0.0.1:8080",
		Secret:   "123",
	})
	if len(l.GetQRCode()) <= 0 {
		t.Error("Get info qrcode failed!")
	}
}
