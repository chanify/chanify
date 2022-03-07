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
	ctx.JSON(http.StatusOK, gin.H{})
}
