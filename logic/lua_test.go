package logic

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

func TestJsonDecode(t *testing.T) {
	l := lua.NewState()
	defer l.Close()
	initLua(l)
	if err := l.DoString(`local json=require "json";return json.decode("123")`); err != nil {
		t.Error("Run json decode value failed:", err)
	}
	if lv := l.Get(-1); lv.String() != "123" {
		t.Error("Do json decode value failed")
	}
	l.Pop(1)
	if err := l.DoString(`local json=require "json";return json.decode("[123,456,\"xyz\",{\"a\":true}]")`); err != nil {
		t.Error("Run json decode array failed:", err)
	}
	if tbl := l.Get(-1).(*lua.LTable); tbl.Len() != 4 {
		t.Error("Do json decode array failed")
	}
	l.Pop(1)
	if err := l.DoString(`local json=require "json";return json.decode("{")`); err != nil {
		t.Error("Run json decode failed:", err)
	}
	if lv := l.Get(-1); lv != lua.LNil {
		t.Error("Do json decode failed")
	}
}

func TestJsonEncode(t *testing.T) {
	l := lua.NewState()
	defer l.Close()
	initLua(l)
	if err := l.DoString(`local json=require "json";return json.encode(123)`); err != nil {
		t.Error("Run json encode value failed:", err)
	}
	if lv := l.Get(-1); lv.String() != "123" {
		t.Error("Do json encode value failed")
	}
	l.Pop(1)
	if err := l.DoString(`local json=require "json";return json.encode({123,456,"xyz",{a=true},json.encode})`); err != nil {
		t.Error("Run json encode array failed:", err)
	}
	if lv := l.Get(-1); lv.String() != `[123,456,"xyz",{"a":true}]` {
		t.Error("Do json encode array failed")
	}
	l.Pop(1)
	if err := l.DoString(`local json=require "json";return json.encode(json.encode)`); err != nil {
		t.Error("Run json encode failed:", err)
	}
	if lv := l.Get(-1); lv.String() != "" {
		t.Error("Do json encode failed", lv)
	}
}

func TestInterface2LValue(t *testing.T) {
	if luaInterface2LValue(nil) != lua.LNil {
		t.Error("Check decode nil failed")
	}
	if luaInterface2LValue(int32(10)) != lua.LNumber(10) {
		t.Error("Check decode int32 failed")
	}
	if luaInterface2LValue(int64(10)) != lua.LNumber(10) {
		t.Error("Check decode int64 failed")
	}
	if luaInterface2LValue(uint32(10)) != lua.LNumber(10) {
		t.Error("Check decode uint32 failed")
	}
	if luaInterface2LValue(uint64(10)) != lua.LNumber(10) {
		t.Error("Check decode uint64 failed")
	}
	if luaInterface2LValue(float32(1.5)) != lua.LNumber(1.5) {
		t.Error("Check decode float32 failed")
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
