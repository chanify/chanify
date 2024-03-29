package core

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/chanify/chanify/logic"
	"github.com/chanify/chanify/model"
	"github.com/gin-gonic/gin"
)

func TestBindBodyJson(t *testing.T) {
	c := New()
	defer c.Close()
	c.Init(&logic.Options{DBUrl: "nosql://?secret=123"}) // nolint: errcheck
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("GET", "/", nil)
	ctx.Request.Body = io.NopCloser(iotest.ErrReader(errors.New("no body")))
	var x int
	if err := c.bindBodyJSON(ctx, &x); err == nil {
		t.Error("Check bind body failed")
	}

	ctx, _ = gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("GET", "/", nil)
	ctx.Request.Header.Set("Content-Type", "application/x-chsec-json")
	ctx.Request.Body = io.NopCloser(strings.NewReader("123"))
	if err := c.bindBodyJSON(ctx, &x); err == nil {
		t.Error("Check bind ecode body failed")
	}
}

func TestVerifyUser(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("GET", "/", nil)
	ctx.Request.Header.Set("CHUserSign", "*****")
	if verifyUser(ctx, "") {
		t.Error("Check verify user failed")
	}
}

func TestVerifyDevice(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("GET", "/", nil)
	ctx.Request.Header.Set("CHDevSign", "*****")
	if verifyDevice(ctx, "") {
		t.Error("Check verify user failed")
	}
}

func TestVerify(t *testing.T) {
	if verifySign("***", []byte{}, []byte{}) {
		t.Fatal("Check verify empty sign failed")
	}
	if verifySign("", []byte{}, []byte{}) {
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
	if _, err := c.parseToken(getToken(ctx)); err != model.ErrInvalidToken {
		t.Fatal("Check get token failed")
	}
}

func TestCreateThumbnail(t *testing.T) {
	dPNG, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAACAQMAAACjTyRkAAAABGdBTUEAALGPC/xhBQAAACBjSFJNAAB6JgAAgIQAAPoAAACA6AAAdTAAAOpgAAA6mAAAF3CculE8AAAABlBMVEWZAAD///+fsNhWAAAAAWJLR0QB/wIt3gAAAAd0SU1FB+UDHRczLl5aCAkAAAAMSURBVAjXY2BgYAAAAAQAASc0JwoAAAAldEVYdGRhdGU6Y3JlYXRlADIwMjEtMDMtMjlUMjM6NTE6NDYrMDA6MDCUDk5dAAAAJXRFWHRkYXRlOm1vZGlmeQAyMDIxLTAzLTI5VDIzOjUxOjQ2KzAwOjAw5VP24QAAAABJRU5ErkJggg==")
	if createThumbnail(dPNG) == nil {
		t.Error("Create png thumbnail failed")
	}
	dGIF, _ := base64.StdEncoding.DecodeString("R0lGODlhAQABAPABAAAAAP///yH5BAAAAAAAIf8LSW1hZ2VNYWdpY2sNZ2FtbWE9MC40NTQ1NQAsAAAAAAEAAQAAAgJMAQA7")
	if createThumbnail(dGIF) == nil {
		t.Error("Create gif thumbnail failed")
	}
	dTIFF, _ := base64.StdEncoding.DecodeString("SUkqAAoAAAD//w8AAAEDAAEAAAABAAAAAQEDAAEAAAABAAAAAgEDAAEAAAAQAAAAAwEDAAEAAAABAAAABgEDAAEAAAABAAAACgEDAAEAAAABAAAAEQEEAAEAAAAIAAAAEgEDAAEAAAABAAAAFQEDAAEAAAABAAAAFgEDAAEAAAABAAAAFwEEAAEAAAACAAAAHAEDAAEAAAABAAAAKQEDAAIAAAAAAAEAPgEFAAIAAAD0AAAAPwEFAAYAAADEAAAAAAAAAIXrUQAAAIAAw/WoAAAAAALNzEwAAAAAAc3MTAAAAIAAzcxMAAAAAAKPwvUAAAAAEDcaoAAAAAACK4cKAAAAIAA=")
	if createThumbnail(dTIFF) == nil {
		t.Error("Create tiff thumbnail failed")
	}
	dJPEG, _ := base64.StdEncoding.DecodeString("/9j/4AAQSkZJRgABAQAAAQABAAD/2wBDAAMCAgICAgMCAgIDAwMDBAYEBAQEBAgGBgUGCQgKCgkICQkKDA8MCgsOCwkJDRENDg8QEBEQCgwSExIQEw8QEBD/2wBDAQMDAwQDBAgEBAgQCwkLEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBD/wAARCAACAAEDAREAAhEBAxEB/8QAFAABAAAAAAAAAAAAAAAAAAAACP/EABQQAQAAAAAAAAAAAAAAAAAAAAD/xAAVAQEBAAAAAAAAAAAAAAAAAAAHCP/EABQRAQAAAAAAAAAAAAAAAAAAAAD/2gAMAwEAAhEDEQA/ABIOllv/2Q==")
	if createThumbnail(dJPEG) == nil {
		t.Error("Create jpeg thumbnail failed")
	}
	dWEBP, _ := base64.StdEncoding.DecodeString("UklGRiQAAABXRUJQVlA4IBgAAAAwAQCdASoBAAEAAgA0JaQAA3AA/vuUAAA=")
	if createThumbnail(dWEBP) == nil {
		t.Error("Create webp thumbnail failed")
	}
}

func TestFileBaseName(t *testing.T) {
	if fileBaseName("..") != "" {
		t.Error("Check file base name failed!")
	}
	if fileBaseName("./123/abc.xyz") != "abc.xyz" {
		t.Error("Get file base name failed!")
	}
}

func TestJsonString(t *testing.T) {
	var data struct {
		A JSONString `json:"a"`
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
