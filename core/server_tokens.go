package core

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/chanify/chanify/model"
	"github.com/gin-gonic/gin"
)

// Management is authenticated by the account signature, never a sending token.
func (c *Core) handleServerTokens(ctx *gin.Context) {
	var params struct {
		User   string `json:"user"`
		Nonce  int64  `json:"nonce"`
		Action string `json:"action"`
		ID     string `json:"id"`
		Name   string `json:"name"`
		Token  string `json:"token"`
	}
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, 16384)
	if c.bindBodyJSON(ctx, &params) != nil {
		ctx.JSON(400, gin.H{"msg": "invalid parameters"})
		return
	}
	user, err := c.logic.GetUser(params.User)
	delta := time.Now().UnixMilli() - params.Nonce
	if err != nil || user.IsServerless() || delta < -300000 || delta > 300000 || !verifyUser(ctx, user.GetPublicKeyString()) {
		ctx.JSON(401, gin.H{"msg": "invalid account signature"})
		return
	}
	records, err := c.logic.ManageServerTokens(params.User, params.Action, params.ID, params.Name, params.Token)
	if err != nil {
		code := 400
		if errors.Is(err, sql.ErrNoRows) {
			code = 404
		} else if !errors.Is(err, model.ErrInvalidToken) && !errors.Is(err, model.ErrNotImplemented) && params.Action == "list" {
			code = 500
		}
		ctx.JSON(code, gin.H{"msg": "token operation failed"})
		return
	}
	ctx.JSON(200, gin.H{"tokens": records})
}
