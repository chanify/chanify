package core

import (
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"net/http"

	cc "github.com/chanify/chanify/crypto"
	"github.com/gin-gonic/gin"
)

var (
	base32Encode = base32.StdEncoding.WithPadding(base32.NoPadding)
	base64Encode = base64.RawURLEncoding
)

func (c *Core) handleBindUser(ctx *gin.Context) {
	var params struct {
		Nonce uint64 `json:"nonce"`
		User  struct {
			Uid string `json:"uid"`
			Key string `json:"key"`
		} `json:"user"`
	}
	if err := ctx.BindJSON(&params); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusBadRequest, "msg": "invalid params"})
		return
	}
	keyData, err := base64Encode.DecodeString(params.User.Key)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusBadRequest, "msg": "invalid public key format"})
		return
	}
	pubKey, err := cc.LoadPublicKey(keyData)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusBadRequest, "msg": "invalid public key"})
		return
	}
	uidData, err := base32Encode.DecodeString(params.User.Uid)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusBadRequest, "msg": "invalid user id"})
		return
	}
	h := sha256.New()
	h.Write(c.info.secret) // nolint: errcheck
	h.Write(uidData)       // nolint: errcheck
	kdata, _ := pubKey.Encrypt(h.Sum(nil))
	key := base64Encode.EncodeToString(kdata)
	ctx.JSON(http.StatusOK, gin.H{"key": key})
}
