package core

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (c *Core) handleBindUser(ctx *gin.Context) {
	var params struct {
		Nonce uint64 `json:"nonce"`
		User  struct {
			Uid string `json:"uid"`
			Key string `json:"key"`
		} `json:"user"`
		Device *struct {
			Uuid      string `json:"uuid"`
			Key       string `json:"key"`
			PushToken string `json:"push-token,omitempty"`
			Sandbox   bool   `json:"sandbox,omitempty"`
		} `json:"device,omitempty"`
	}
	if err := ctx.BindJSON(&params); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusBadRequest, "msg": "invalid params"})
		return
	}
	serverless := (params.Device == nil)
	u, err := c.logic.UpsertUser(params.User.Uid, params.User.Key, serverless)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusBadRequest, "msg": "invalid user id"})
		return
	}
	if serverless {
		log.Println("Bind user:", params.User.Uid)
	} else {
		if err := c.logic.BindDevice(params.User.Uid, params.Device.Uuid, params.Device.Key); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusBadRequest, "msg": "bind user device failed"})
			return
		}
		log.Println("Bind user:", params.User.Uid, "device:", params.Device.Uuid)
		if len(params.Device.PushToken) > 0 {
			c.logic.UpdatePushToken(params.User.Uid, params.Device.Uuid, params.Device.PushToken, params.Device.Sandbox) // nolint:errcheck
		}
	}
	kdata := u.PublicKeyEncrypt(u.SecretKey)
	ctx.JSON(http.StatusOK, gin.H{"key": base64Encode.EncodeToString(kdata)})
}

func (c *Core) handleUnbindUser(ctx *gin.Context) {
	var params struct {
		Nonce    uint64 `json:"nonce"`
		DeviceID string `json:"device"`
		UserID   string `json:"user"`
	}
	if err := ctx.BindJSON(&params); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusBadRequest, "msg": "unbind user device failed"})
		return
	}
	c.logic.UnbindDevice(params.UserID, params.DeviceID) // nolint: errcheck
	ctx.JSON(http.StatusOK, gin.H{
		"uuid": params.DeviceID,
		"uid":  params.UserID,
	})
}
