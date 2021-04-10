package core

import (
	"net/http"

	"github.com/chanify/chanify/model"
	"github.com/gin-gonic/gin"
)

func (c *Core) handleImageDownload(ctx *gin.Context) {
	token, _ := c.getToken(ctx)
	c.downloadImageFile(ctx, token)
}

func (c *Core) handleFileDownload(ctx *gin.Context) {
	token, _ := c.getToken(ctx)
	c.downloadFile(ctx, token)
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
	data, err := c.logic.LoadFile("images", fname)
	if err != nil {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}
	ctx.Data(http.StatusOK, parseImageContentType(data), data)
}

func (c *Core) downloadFile(ctx *gin.Context, token *model.Token) {
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
	data, err := c.logic.LoadFile("files", fname)
	if err != nil {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}
	ctx.Data(http.StatusOK, "application/octet-stream", data)
}
