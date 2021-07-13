package core

import (
	"net/http/httptest"
	"testing"
	"time"

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

func TestParseTimestamp(t *testing.T) {
	tm := time.Unix(1620000000, 123000000)
	if !tm.Equal(*parseTimestamp("1620000000123")) {
		t.Error("Parse unix time failed")
	}
	if !tm.Equal(*parseTimestamp(tm.Format(time.RFC3339Nano))) {
		t.Error("Parse string time nano failed")
	}
	tm2 := time.Unix(1620000000, 0)
	if !tm2.Equal(*parseTimestamp(tm.Format(time.RFC3339))) {
		t.Error("Parse string time failed")
	}
	if !tm.Equal(*parseTimestamp(1620000000123)) {
		t.Error("Parse int time failed")
	}
	if !tm.Equal(*parseTimestamp(uint(1620000000123))) {
		t.Error("Parse uint time failed")
	}
	if !tm.Equal(*parseTimestamp(int64(1620000000123))) {
		t.Error("Parse int64 time failed")
	}
	if !tm.Equal(*parseTimestamp(uint64(1620000000123))) {
		t.Error("Parse uint64 failed")
	}
}
