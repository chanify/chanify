package core

import (
	"net/http"

	"github.com/chanify/chanify/model"
	"github.com/gin-gonic/gin"
)

const (
	pngHeader = "\x89PNG\r\n\x1a\n"
)

func (c *Core) handleImageFile(ctx *gin.Context) {
	token, _ := c.getToken(ctx)
	c.downloadImageFile(ctx, token)
}

func (c *Core) downloadImageFile(ctx *gin.Context, token *model.Token) {
	if token == nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if !token.VerifyDataHash([]byte(ctx.Request.URL.Path)) {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	fname := ctx.Param("fname")
	if len(fname) <= 0 {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	data, err := c.logic.LoadImageFile(fname)
	if err != nil {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}
	ctx.Data(http.StatusOK, parseImageContentType(data), data)
}
