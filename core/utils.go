package core

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"

	"github.com/chanify/chanify/crypto"
	"github.com/chanify/chanify/model"
	"github.com/gin-gonic/gin"
)

func ValidateUser(ctx *gin.Context, key string) bool {
	sign, err := base64Encode.DecodeString(ctx.GetHeader("CHUserSign"))
	if err != nil {
		return false
	}
	data, _ := ctx.Get(gin.BodyBytesKey)
	return ValidateSign(key, sign, data.([]byte))
}

func ValidateDevice(ctx *gin.Context, key string) bool {
	sign, err := base64Encode.DecodeString(ctx.GetHeader("CHDevSign"))
	if err != nil {
		return false
	}
	data, _ := ctx.Get(gin.BodyBytesKey)
	return ValidateSign(key, sign, data.([]byte))
}

func ValidateSign(key string, sign []byte, data []byte) bool {
	kd, err := base64Encode.DecodeString(key)
	if err != nil {
		return false
	}
	pk, err := crypto.LoadPublicKey(kd)
	if err != nil {
		return false
	}
	return pk.Verify(data, sign)
}

func NewAESGCM(key []byte) (cipher.AEAD, error) {
	if len(key) < 32 {
		return nil, errors.New("invalid key")
	}
	block, _ := aes.NewCipher(key[:32])
	return cipher.NewGCM(block)
}

func getToken(ctx *gin.Context) (*model.Token, error) {
	token := ctx.GetHeader("token")
	if len(token) <= 0 {
		token = ctx.Query("token")
		if len(token) <= 0 {
			token = ctx.Param("token")
			if len(token) > 0 && token[0] == '/' {
				token = token[1:]
			}
		}
	}
	return model.ParseToken(token)
}
