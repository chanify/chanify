package core

import (
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

func TestNewAESGCM(t *testing.T) {
	if _, err := NewAESGCM([]byte{}); err == nil {
		t.Error("Check new aes gcm failed")
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
