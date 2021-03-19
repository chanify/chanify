package core

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"

	"github.com/chanify/chanify/model"
	"github.com/gin-gonic/gin"
)

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
