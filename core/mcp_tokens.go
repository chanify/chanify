package core

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// MCP credential management uses an account signature, never MCP/send bearers.
func (c *Core) handleMCPTokens(ctx *gin.Context) {
	var p struct {
		User    string `json:"user"`
		Nonce   int64  `json:"nonce"`
		Action  string `json:"action"`
		ID      string `json:"id"`
		Name    string `json:"name"`
		Expires int64  `json:"expires"`
	}
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, 16384)
	if c.bindBodyJSON(ctx, &p) != nil {
		ctx.JSON(400, gin.H{"msg": "invalid parameters"})
		return
	}
	user, err := c.logic.GetUser(p.User)
	delta := time.Now().UnixMilli() - p.Nonce
	if err != nil || user.IsServerless() || delta < -300000 || delta > 300000 || !verifyUser(ctx, user.GetPublicKeyString()) {
		ctx.JSON(401, gin.H{"msg": "invalid account signature"})
		return
	}
	switch p.Action {
	case "create":
		r, raw, err := c.logic.CreateMCPToken(p.User, p.Name, p.Expires)
		if err != nil {
			ctx.JSON(400, gin.H{"msg": "could not create MCP credential"})
			return
		}
		ctx.JSON(200, gin.H{"credential": r, "token": raw})
		return
	case "rename", "revoke":
		if err := c.logic.UpdateMCPToken(p.User, p.ID, p.Name, p.Action == "revoke"); err != nil {
			ctx.JSON(400, gin.H{"msg": "MCP credential operation failed"})
			return
		}
	case "list":
	default:
		ctx.JSON(400, gin.H{"msg": "invalid action"})
		return
	}
	records, err := c.logic.ListMCPTokens(p.User)
	if err != nil {
		ctx.JSON(500, gin.H{"msg": "could not list MCP credentials"})
		return
	}
	ctx.JSON(200, gin.H{"tokens": records})
}
