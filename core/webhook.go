package core

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/chanify/chanify/model"
	"github.com/gin-gonic/gin"
	lua "github.com/yuin/gopher-lua"
)

const coreKey = "_chanify/http/core"

func (c *Core) handlePostWebhook(ctx *gin.Context) {
	name := strings.ToLower(ctx.Param("name"))
	webhook, err := c.logic.GetWebhook(name)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"res": http.StatusNotFound, "msg": "no webhook found"})
		return
	}
	ctx.Set(coreKey, c)
	if ctx.Request.Body != nil {
		if body, err := ioutil.ReadAll(ctx.Request.Body); err == nil {
			ctx.Set(gin.BodyBytesKey, body)
		}
	}
	l := lua.NewState()
	defer l.Close()
	initHttpLua(l, ctx)
	if err := webhook.DoCall(l); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusBadRequest, "msg": err.Error()})
		return
	}
	code, ctype, data := getHttpLuaReturn(l)
	ctx.DataFromReader(code, int64(len(data)), ctype, strings.NewReader(data), map[string]string{})
}

func initHttpLua(l *lua.LState, ctx *gin.Context) {
	l.SetField(l.NewTypeMetatable("Request"), "__index", l.SetFuncs(l.NewTable(), luaRequestMethods))

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
	"request": luaContextGetRequest,
	"send":    luaContextSend,
}

var luaRequestMethods = map[string]lua.LGFunction{
	"token":  luaContextGetToken,
	"url":    luaContextGetUrl,
	"body":   luaContextGetBody,
	"query":  luaContextGetQuery,
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

func luaContextGetRequest(l *lua.LState) int {
	ctx := luaCheckContext(l)
	lc := l.NewUserData()
	lc.Value = ctx
	lc.Metatable = l.GetTypeMetatable("Request")
	l.Push(lc)
	return 1
}

func luaContextSend(l *lua.LState) int {
	ctx := luaCheckContext(l)
	cc, ok := ctx.Get(coreKey)
	if !ok {
		l.Push(lua.LString("error: unknown error"))
		return 1
	}
	c := cc.(*Core)
	text := l.CheckString(2)
	if len(text) <= 0 {
		l.Push(lua.LString("error: invalid message"))
		return 1
	}
	token, err := c.parseToken(getToken(ctx))
	if err != nil {
		l.Push(lua.LString("error: invalid token"))
		return 1
	}
	msg := model.NewMessage(token)
	msg, err = c.makeTextContent(msg, text, ctx.Query("title"), ctx.Query("copy"), ctx.Query("autocopy"), ctx.QueryArray("action"))
	if err != nil {
		l.Push(lua.LString("error: too large text content"))
		return 1
	}
	ret := c.sendMsg(ctx, token, msg.SoundName(ctx.Query("sound")).SetPriority(parsePriority(ctx.Query("priority"))).SetInterruptionLevel(ctx.Query("interruption-level")))
	l.Push(lua.LString(ret))
	return 1
}

func luaContextGetToken(l *lua.LState) int {
	ctx := luaCheckContext(l)
	l.Push(lua.LString(getToken(ctx)))
	return 1
}

func luaContextGetUrl(l *lua.LState) int {
	ctx := luaCheckContext(l)
	l.Push(lua.LString(ctx.Request.URL.String()))
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

func luaContextGetQuery(l *lua.LState) int {
	ctx := luaCheckContext(l)
	if v, ok := ctx.GetQuery(l.CheckString(2)); ok {
		l.Push(lua.LString(v))
	} else {
		l.Push(lua.LNil)
	}
	return 1
}

func luaContextGetHeader(l *lua.LState) int {
	ctx := luaCheckContext(l)
	key := strings.ToLower(l.CheckString(2))
	switch key {
	default:
		l.Push(lua.LString(ctx.GetHeader(key)))
	case "host":
		l.Push(lua.LString(ctx.Request.Host))
	case "user-agent":
		l.Push(lua.LString(ctx.Request.UserAgent()))
	case "content-length":
		l.Push(lua.LNumber(ctx.Request.ContentLength))
	}
	return 1
}
