package core

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	lua "github.com/yuin/gopher-lua"
)

func (c *Core) handlePostWebhook(ctx *gin.Context) {
	name := strings.ToLower(ctx.Param("name"))
	whProto, err := c.logic.GetWebhook(name)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"res": http.StatusNotFound, "msg": "no webhook found"})
		return
	}
	if ctx.Request.Body != nil {
		if body, err := ioutil.ReadAll(ctx.Request.Body); err == nil {
			ctx.Set(gin.BodyBytesKey, body)
		}
	}

	l := lua.NewState()
	defer l.Close()
	initHttpLua(l, ctx)
	whFunc := l.NewFunctionFromProto(whProto)
	l.Push(whFunc)
	if err := l.PCall(0, lua.MultRet, nil); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusBadRequest, "msg": err.Error()})
		return
	}
	code, ctype, data := getHttpLuaReturn(l)
	ctx.DataFromReader(code, int64(len(data)), ctype, strings.NewReader(data), map[string]string{})
}

func initHttpLua(l *lua.LState, ctx *gin.Context) {
	initLua(l)

	mt := l.NewTypeMetatable("Context")
	l.SetField(mt, "__index", l.SetFuncs(l.NewTable(), luaContextMethods))

	lc := l.NewUserData()
	lc.Value = ctx
	lc.Metatable = mt
	l.SetGlobal("ctx", lc)
}

func getHttpLuaReturn(l *lua.LState) (int, string, string) {
	code := http.StatusOK
	ctype := "text/plain; charset=utf-8"
	value := ""
	lv := l.Get(-1)
	if lv != lua.LNil {
		if val, ok := lv.(lua.LNumber); ok {
			code = int(val)
		} else {
			value = lv.String()
			lv = l.Get(-2)
			if lv != lua.LNil {
				if val, ok := lv.(lua.LNumber); ok {
					code = int(val)
				} else {
					ctype = lv.String()
					lv = l.Get(-3)
					if lv != lua.LNil {
						if val, ok := lv.(lua.LNumber); ok {
							code = int(val)
						}
					}
				}
			}
		}
	}
	return code, ctype, value
}

var luaContextMethods = map[string]lua.LGFunction{
	"token":  luaContextGetToken,
	"body":   luaContextGetBody,
	"header": luaContextGetHeader,
}

func luaCheckContext(l *lua.LState) *gin.Context {
	ud := l.CheckUserData(1)
	if v, ok := ud.Value.(*gin.Context); ok {
		return v
	}
	l.ArgError(1, "http context expected")
	return nil
}

func luaContextGetToken(l *lua.LState) int {
	ctx := luaCheckContext(l)
	l.Push(lua.LString(getToken(ctx)))
	return 1
}

func luaContextGetBody(l *lua.LState) int {
	ctx := luaCheckContext(l)
	if cb, ok := ctx.Get(gin.BodyBytesKey); ok {
		if cbb, ok := cb.([]byte); ok {
			l.Push(lua.LString(cbb))
			return 1
		}
	}
	l.Push(lua.LString(""))
	return 1
}

func luaContextGetHeader(l *lua.LState) int {
	ctx := luaCheckContext(l)
	l.Push(lua.LString(ctx.GetHeader(l.CheckString(2))))
	return 1
}
