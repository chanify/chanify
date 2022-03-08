package core

import (
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
	l := lua.NewState()
	defer l.Close()
	whFunc := l.NewFunctionFromProto(whProto)
	l.Push(whFunc)
	if err := l.PCall(0, lua.MultRet, nil); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusBadRequest, "msg": err.Error()})
		return
	}
	code, ctype, data := getHttpLuaReturn(l)
	ctx.DataFromReader(code, int64(len(data)), ctype, strings.NewReader(data), map[string]string{})
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
