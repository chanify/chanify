package core

import (
	"encoding/json"
	"fmt"
	"io"
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
		if body, err := io.ReadAll(ctx.Request.Body); err == nil {
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

type luaSendContext struct {
	msg string
}

func luaContextSend(l *lua.LState) int {
	ctx := luaCheckContext(l)
	cc, ok := ctx.Get(coreKey)
	if !ok {
		l.Push(lua.LString(`{"res":500,"msg":"unknown error"}`))
		return 1
	}
	c := cc.(*Core)
	var text string
	opts := l.NewTable()
	switch args := l.Get(2).(type) {
	case lua.LString:
		text = string(args)
	case *lua.LTable:
		opts = args
		text = opts.RawGetString("text").String()
	}
	if len(text) <= 0 {
		l.Push(lua.LString(`{"res":204,"msg":"no message content"}`))
		return 1
	}
	tk := luaGetOptsString(opts, "token")
	if len(tk) <= 0 {
		tk = getToken(ctx)
	}
	token, err := c.parseToken(tk)
	if err != nil {
		l.Push(lua.LString(`{"res":401,"msg":"invalid token"}`))
		return 1
	}
	msg := model.NewMessage(token)
	msg, err = c.makeTextContent(msg, text, luaGetOptsString(opts, "title"), luaGetOptsString(opts, "copy"), luaGetOptsString(opts, "autocopy"), luaGetOptsArray(opts, "action"))
	if err != nil {
		l.Push(lua.LString(`{"res":413,"msg":"too large text content"}`))
		return 1
	}
	lc := &luaSendContext{}
	c.sendMsg(lc, token, msg.SoundName(luaGetOptsString(opts, "sound")).SetPriority(parsePriority(luaGetOptsString(opts, "priority"))).SetInterruptionLevel(luaGetOptsString(opts, "interruption-level")))
	l.Push(lua.LString(lc.String()))
	return 1
}

func (l *luaSendContext) String() string {
	return l.msg
}

func (l *luaSendContext) JSON(code int, obj interface{}) {
	if val, err := json.Marshal(obj); err == nil {
		l.msg = string(val)
	}
}

func (l *luaSendContext) DataFromReader(code int, contentLength int64, contentType string, reader io.Reader, extraHeaders map[string]string) {
	if val, err := io.ReadAll(reader); err == nil {
		l.msg = string(val)
	}
}

func luaGetOptsString(opts *lua.LTable, key string) string {
	val := opts.RawGetString(key)
	switch v := val.(type) {
	default:
		return ""
	case lua.LString:
		return string(v)
	case lua.LBool, lua.LNumber:
		return fmt.Sprint(v)
	}
}

func luaGetOptsArray(opts *lua.LTable, key string) []string {
	actions := []string{}
	val := opts.RawGetString(key)
	if v, ok := val.(*lua.LTable); ok {
		v.ForEach(func(idx, value lua.LValue) {
			actions = append(actions, value.String())
		})
	}
	return actions
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
