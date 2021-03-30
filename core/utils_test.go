package core

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chanify/chanify/model"
	"github.com/gin-gonic/gin"
)

func TestVerifyUser(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("GET", "/", nil)
	ctx.Request.Header.Set("CHUserSign", "*****")
	if VerifyUser(ctx, "") {
		t.Error("Check verify user failed")
	}
}

func TestVerifyDevice(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("GET", "/", nil)
	ctx.Request.Header.Set("CHDevSign", "*****")
	if VerifyDevice(ctx, "") {
		t.Error("Check verify user failed")
	}
}

func TestVerify(t *testing.T) {
	if VerifySign("***", []byte{}, []byte{}) {
		t.Fatal("Check verify empty sign failed")
	}
	if VerifySign("", []byte{}, []byte{}) {
		t.Fatal("Check verify invalid key sign failed")
	}
}

func TestGetToken(t *testing.T) {
	c := New()
	defer c.Close()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("GET", "/", nil)
	ctx.Params = []gin.Param{{Key: "token", Value: "/EgMxMjMiBGNoYW4qBU1GUkdH..c2lnbg"}}
	if _, err := c.getToken(ctx); err != model.ErrInvalidToken {
		t.Fatal("Check get token failed")
	}
}

func TestParseImageContentType(t *testing.T) {
	if parseImageContentType([]byte{137, 80, 78, 71, 13, 10, 26, 10, 0}) != "image/png" {
		t.Fatal("Parse png header failed")
	}
}

func TestCreateThumbnail(t *testing.T) {
	d1, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAACAQMAAACjTyRkAAAABGdBTUEAALGPC/xhBQAAACBjSFJNAAB6JgAAgIQAAPoAAACA6AAAdTAAAOpgAAA6mAAAF3CculE8AAAABlBMVEWZAAD///+fsNhWAAAAAWJLR0QB/wIt3gAAAAd0SU1FB+UDHRczLl5aCAkAAAAMSURBVAjXY2BgYAAAAAQAASc0JwoAAAAldEVYdGRhdGU6Y3JlYXRlADIwMjEtMDMtMjlUMjM6NTE6NDYrMDA6MDCUDk5dAAAAJXRFWHRkYXRlOm1vZGlmeQAyMDIxLTAzLTI5VDIzOjUxOjQ2KzAwOjAw5VP24QAAAABJRU5ErkJggg==")
	tb1 := CreateThumbnail(d1)
	if tb1 == nil {
		t.Error("Create png thumbnail failed")
	}
	d2, _ := base64.StdEncoding.DecodeString("/9j/4AAQSkZJRgABAQAAAQABAAD/2wBDAAMCAgICAgMCAgIDAwMDBAYEBAQEBAgGBgUGCQgKCgkICQkKDA8MCgsOCwkJDRENDg8QEBEQCgwSExIQEw8QEBD/2wBDAQMDAwQDBAgEBAgQCwkLEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBD/wAARCAACAAEDAREAAhEBAxEB/8QAFAABAAAAAAAAAAAAAAAAAAAACP/EABQQAQAAAAAAAAAAAAAAAAAAAAD/xAAVAQEBAAAAAAAAAAAAAAAAAAAHCP/EABQRAQAAAAAAAAAAAAAAAAAAAAD/2gAMAwEAAhEDEQA/ABIOllv/2Q==")
	tb2 := CreateThumbnail(d2)
	if tb2 == nil {
		t.Error("Create jpeg thumbnail failed")
	}
}

func TestJsonString(t *testing.T) {
	var data struct {
		A JsonString `json:"a"`
	}
	if err := json.Unmarshal([]byte(`{"a":"false"}`), &data); err != nil {
		t.Fatal("Unmarshal json failed", err)
	}
	if len(data.A) > 0 {
		t.Fatal("Check unmarshal json failed")
	}
	if err := json.Unmarshal([]byte(`{"a":"abc"}`), &data); err != nil {
		t.Fatal("Unmarshal json failed", err)
	}
	if data.A != "abc" {
		t.Fatal("Check unmarshal json failed")
	}
}
