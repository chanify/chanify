package core

import (
	"net/http/httptest"
	"testing"

	"github.com/chanify/chanify/logic"
	"github.com/gin-gonic/gin"
)

func TestMsgImageParam(t *testing.T) {
	c := New()
	defer c.Close()
	c.Init(&logic.Options{DBUrl: "sqlite://?mode=memory"}) // nolint: errcheck
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	msg := &MsgParam{}
	if _, err := msg.ParseImage(c, ctx); err != nil {
		t.Error("Parse image params failed:", err)
	}
}

func TestMsgAudioParam(t *testing.T) {
	c := New()
	defer c.Close()
	c.Init(&logic.Options{DBUrl: "sqlite://?mode=memory"}) // nolint: errcheck
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	msg := &MsgParam{}
	if _, err := msg.ParseAudio(c, ctx); err != nil {
		t.Error("Parse audio params failed:", err)
	}
}
