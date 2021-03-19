package core

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

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
