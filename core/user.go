package core

import (
	"log"
	"net/http"

	"github.com/chanify/chanify/crypto"
	"github.com/chanify/chanify/logic"
	"github.com/gin-gonic/gin"
)

func (c *Core) handleBindUser(ctx *gin.Context) {
	var params struct {
		Nonce uint64 `json:"nonce"`
		User  struct {
			UID string `json:"uid"`
			Key string `json:"key"`
		} `json:"user"`
		Device *struct {
			UUID      string `json:"uuid"`
			Key       string `json:"key"`
			PushToken string `json:"push-token,omitempty"`
			Sandbox   bool   `json:"sandbox,omitempty"`
			Type      int    `json:"type,omitempty"`
		} `json:"device,omitempty"`
	}
	if err := c.bindBodyJSON(ctx, &params); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusBadRequest, "msg": "invalid params"})
		return
	}
	if !verifyUser(ctx, params.User.Key) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"res": http.StatusUnauthorized, "msg": "invalid user sign"})
		return
	}
	if params.Device != nil && !verifyDevice(ctx, params.Device.Key) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"res": http.StatusUnauthorized, "msg": "invalid device sign"})
		return
	}
	serverless := (params.Device == nil)
	u, err := c.logic.UpsertUser(params.User.UID, params.User.Key, serverless)
	if err != nil {
		if err == logic.ErrSystemLimited {
			ctx.JSON(http.StatusNotAcceptable, gin.H{"res": http.StatusNotAcceptable, "msg": "system limited"})
		} else {
			ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusBadRequest, "msg": "invalid user id"})
		}
		return
	}
	if serverless {
		log.Println("Bind user:", params.User.UID)
	} else {
		if err := c.logic.BindDevice(params.User.UID, params.Device.UUID, params.Device.Key, params.Device.Type); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusBadRequest, "msg": "bind user device failed"})
			return
		}
		log.Println("Bind user:", params.User.UID, "device:", params.Device.UUID)
		if len(params.Device.PushToken) > 0 {
			c.logic.UpdatePushToken(params.User.UID, params.Device.UUID, params.Device.PushToken, params.Device.Sandbox) // nolint:errcheck
		}
	}
	kdata := u.PublicKeyEncrypt(u.SecretKey)
	ctx.JSON(http.StatusOK, gin.H{"key": crypto.Base64Encode.EncodeToString(kdata)})
}

func (c *Core) handleUnbindUser(ctx *gin.Context) {
	var params struct {
		Nonce    uint64 `json:"nonce"`
		DeviceID string `json:"device"`
		UserID   string `json:"user"`
	}
	if err := c.bindBodyJSON(ctx, &params); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusBadRequest, "msg": "unbind user device failed"})
		return
	}
	u, err := c.logic.GetUser(params.UserID)
	if err == nil && !u.IsServerless() && !verifyUser(ctx, u.GetPublicKeyString()) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"res": http.StatusUnauthorized, "msg": "invalid user sign"})
		return
	}
	c.logic.UnbindDevice(params.UserID, params.DeviceID) // nolint: errcheck
	ctx.JSON(http.StatusOK, gin.H{
		"uuid": params.DeviceID,
		"uid":  params.UserID,
	})
}
