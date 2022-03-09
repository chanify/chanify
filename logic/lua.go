package logic

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"hash"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

var luaMods = map[string]map[string]lua.LGFunction{
	"hex": {
		"decode": luaHEXDecode,
		"encode": luaHEXEncode,
	},
	"json": {
		"decode": luaJsonDecode,
		"encode": luaJsonEncode,
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

func luaJsonDecode(l *lua.LState) int {
	var value interface{}
	if err := json.Unmarshal([]byte(l.CheckString(1)), &value); err != nil {
		l.Push(lua.LNil)
		return 1
	}
	l.Push(luaInterface2LValue(value))
	return 1
}

func luaJsonEncode(l *lua.LState) int {
	value := luaLValue2Interface(l.CheckAny(1))
	if value == nil {
		l.Push(lua.LString(""))
		return 1
	}
	data, _ := json.Marshal(value) // nolint:errcheck
	l.Push(lua.LString(data))
	return 1
}

func luaInterface2LValue(v interface{}) lua.LValue {
	switch val := v.(type) {
	default:
		return lua.LNil
	case int32:
		return lua.LNumber(val)
	case int64:
		return lua.LNumber(val)
	case uint32:
		return lua.LNumber(val)
	case uint64:
		return lua.LNumber(val)
	case float32:
		return lua.LNumber(val)
	case float64:
		return lua.LNumber(val)
	case bool:
		return lua.LBool(val)
	case string:
		return lua.LString(val)
	case []interface{}:
		a := lua.LTable{}
		for _, vv := range val {
			if vval := luaInterface2LValue(vv); vval != nil {
				a.Append(vval)
			}
		}
		return &a
	case map[string]interface{}:
		m := lua.LTable{}
		for kk, vv := range val {
			if vval := luaInterface2LValue(vv); vval != nil {
				m.RawSetString(kk, vval)
			}
		}
		return &m
	}
}

func luaLValue2Interface(v lua.LValue) interface{} {
	switch val := v.(type) {
	default:
		return nil
	case lua.LBool:
		return bool(val)
	case lua.LNumber:
		return float64(val)
	case lua.LString:
		return string(val)
	case *lua.LTable:
		a := []interface{}{}
		var m map[string]interface{}
		val.ForEach(func(key, value lua.LValue) {
			vv := luaLValue2Interface(value)
			if vv != nil {
				if k, ok := key.(lua.LString); ok {
					if m == nil {
						m = map[string]interface{}{}
					}
					m[string(k)] = vv
				} else {
					a = append(a, vv)
				}
			}
		})
		if m != nil {
			return m
		}
		return a
	}
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
