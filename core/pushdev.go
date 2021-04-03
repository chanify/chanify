package core

import (
	"net/http"

	"github.com/chanify/chanify/crypto"
	"github.com/gin-gonic/gin"
)

func (c *Core) handleUpdatePushToken(ctx *gin.Context) {
	var params struct {
		Nonce    uint64 `json:"nonce"`
		DeviceID string `json:"device"`
		UserID   string `json:"user"`
		Token    string `json:"token"`
		Sandbox  bool   `json:"sandbox,omitempty"`
	}
	if err := c.BindBodyJson(ctx, &params); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusBadRequest, "msg": "invalid params"})
		return
	}
	u, err := c.logic.GetUser(params.UserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusBadRequest, "msg": "invalid user id"})
		return
	}
	if u.IsServerless() {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusBadRequest, "msg": "invalid user mode"})
		return
	}
	if !VerifyUser(ctx, u.GetPublicKeyString()) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"res": http.StatusUnauthorized, "msg": "invalid user sign"})
		return
	}
	dev, err := c.logic.GetDeviceKey(params.DeviceID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusBadRequest, "msg": "invalid device id"})
		return
	}
	if !VerifyDevice(ctx, crypto.Base64Encode.EncodeToString(dev)) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"res": http.StatusUnauthorized, "msg": "invalid device sign"})
		return
	}
	if err := c.logic.UpdatePushToken(params.UserID, params.DeviceID, params.Token, params.Sandbox); err != nil {
		ctx.JSON(http.StatusConflict, gin.H{"res": http.StatusConflict, "msg": "update push token failed"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"uuid": params.DeviceID,
		"uid":  params.UserID,
	})
}
