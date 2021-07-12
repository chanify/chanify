package core

import (
	"net/http/httptest"
	"testing"

	"github.com/chanify/chanify/logic"
	"github.com/chanify/chanify/model"
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

func TestParseTimeContentItems(t *testing.T) {
	items := map[string]interface{}{}
	items["key1"] = 123
	items["key2"] = int64(1234)
	items["key3"] = float32(456)
	if len(parseTimeContentItems(items)) != len(items) {
		t.Error("Parse time content failed!")
	}
}

func TestTryFormMap(t *testing.T) {
	items := []*model.MsgTimeItem{
		{Name: "123", Value: 456},
	}
	if len(tryFormMap(nil, "test", items)) != len(items) {
		t.Error("Try form map failed!")
	}
}
