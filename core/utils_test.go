package core

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestValidateUser(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("GET", "/", nil)
	ctx.Request.Header.Set("CHUserSign", "*****")
	if ValidateUser(ctx, "") {
		t.Error("Check validate user failed")
	}
}

func TestValidateDevice(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("GET", "/", nil)
	ctx.Request.Header.Set("CHDevSign", "*****")
	if ValidateDevice(ctx, "") {
		t.Error("Check validate user failed")
	}
}

func TestValidate(t *testing.T) {
	if ValidateSign("***", []byte{}, []byte{}) {
		t.Fatal("Check validate empty sign failed")
	}
	if ValidateSign("", []byte{}, []byte{}) {
		t.Fatal("Check validate invalid key sign failed")
	}
}

func TestNewAESGCM(t *testing.T) {
	if _, err := NewAESGCM([]byte{}); err == nil {
		t.Error("Check new aes gcm failed")
	}
}

func TestGetToken(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("GET", "/", nil)
	ctx.Params = []gin.Param{{Key: "token", Value: "/EgMxMjMiBGNoYW4qBU1GUkdH.c2lnbg"}}
	if _, err := getToken(ctx); err != nil {
		t.Fatal("Check get token failed")
	}
}
