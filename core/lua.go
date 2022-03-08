package core

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"hash"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

var luaMods = map[string]map[string]lua.LGFunction{
	"hex": {
		"decode": luaHEXDecode,
		"encode": luaHEXEncode,
	},
	"crypto": {
		"equal": luaCryptoEqual,
		"hmac":  luaCryptoHmac,
	},
}

func luaHEXDecode(l *lua.LState) int {
	data, err := hex.DecodeString(l.CheckString(1))
	if err != nil {
		l.Push(lua.LString(""))
		return 1
	}
	l.Push(lua.LString(data))
	return 1
}

func luaHEXEncode(l *lua.LState) int {
	l.Push(lua.LString(hex.EncodeToString([]byte(l.CheckString(1)))))
	return 1
}

func luaCryptoEqual(l *lua.LState) int {
	if subtle.ConstantTimeCompare([]byte(l.CheckString(1)), []byte(l.CheckString(2))) == 1 {
		l.Push(lua.LTrue)
		return 1
	}
	l.Push(lua.LFalse)
	return 1
}

func luaCryptoHmac(l *lua.LState) int {
	var h func() hash.Hash
	switch strings.ToLower(l.CheckString(1)) {
	default:
		l.Push(lua.LString(""))
		return 1
	case "md5":
		h = md5.New
	case "sha1":
		h = sha1.New
	case "sha256":
		h = sha256.New
	}
	mac := hmac.New(h, []byte(l.CheckString(2)))
	mac.Write([]byte(l.CheckString(3)))
	l.Push(lua.LString(mac.Sum(nil)))
	return 1
}

func initLua(l *lua.LState) {
	for v, m := range luaMods {
		mod := m
		l.PreloadModule(v, func(l *lua.LState) int {
			mod := l.SetFuncs(l.NewTable(), mod)
			l.Push(mod)
			return 1
		})
	}
}
