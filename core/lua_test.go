package core

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestLuaHEX(t *testing.T) {
	l := lua.NewState()
	defer l.Close()
	initLua(l)
	if err := l.DoString(`local hex=require "hex";return hex.encode("abc")`); err != nil {
		t.Error("Run encode failed:", err)
	}
	if lv := l.Get(-1); lv.String() != "616263" {
		t.Error("Do encode failed")
	}
	l.Pop(1)
	if err := l.DoString(`local hex=require "hex";return hex.decode("616263")`); err != nil {
		t.Error("Run encode failed:", err)
	}
	if lv := l.Get(-1); lv.String() != "abc" {
		t.Error("Do encode failed")
	}
	l.Pop(1)
	if err := l.DoString(`local hex=require "hex";return hex.decode("----")`); err != nil {
		t.Error("Check run encode failed:", err)
	}
	if lv := l.Get(-1); lv.String() != "" {
		t.Error("Check do encode failed")
	}
}

func TestLuaCryptoEqual(t *testing.T) {
	l := lua.NewState()
	defer l.Close()
	initLua(l)
	if err := l.DoString(`local c=require "crypto";return c.equal("123","123")`); err != nil {
		t.Error("Run crypto equal failed:", err)
	}
	if lv := l.Get(-1); lv.String() != "true" {
		t.Error("Do crypto equal failed")
	}
	l.Pop(1)
	if err := l.DoString(`local c=require "crypto";return c.equal("123","abc")`); err != nil {
		t.Error("Run crypto not equal failed:", err)
	}
	if lv := l.Get(-1); lv.String() != "false" {
		t.Error("Do crypto not equal failed")
	}
}

func TestLuaCryptoHMac(t *testing.T) {
	l := lua.NewState()
	defer l.Close()
	initLua(l)
	if err := l.DoString(`local c=require "crypto";local h=require "hex";return h.encode(c.hmac("md5","123","abc"))`); err != nil {
		t.Error("Run crypto hmac md5 failed:", err)
	}
	if lv := l.Get(-1); lv.String() != "ffb7c0fc166f7ca075dfa04d59aed232" {
		t.Error("Do crypto hmac md5 failed")
	}
	l.Pop(1)
	if err := l.DoString(`local c=require "crypto";local h=require "hex";return h.encode(c.hmac("sha1","123","abc"))`); err != nil {
		t.Error("Run crypto hmac sha1 failed:", err)
	}
	if lv := l.Get(-1); lv.String() != "540b0c53d4925837bd92b3f71abe7a9d70b676c4" {
		t.Error("Do crypto hmac sha1 failed")
	}
	l.Pop(1)
	if err := l.DoString(`local c=require "crypto";local h=require "hex";return h.encode(c.hmac("sha256","123","abc"))`); err != nil {
		t.Error("Run crypto hmac sha256 failed:", err)
	}
	if lv := l.Get(-1); lv.String() != "8f16771f9f8851b26f4d460fa17de93e2711c7e51337cb8a608a0f81e1c1b6ae" {
		t.Error("Do crypto hmac sha256 failed")
	}
	l.Pop(1)
	if err := l.DoString(`local c=require "crypto";return c.hmac("xyz","123","abc")`); err != nil {
		t.Error("Run crypto invalid hmac failed:", err)
	}
	if lv := l.Get(-1); lv.String() != "" {
		t.Error("Do crypto invalid hmac failed")
	}
}
