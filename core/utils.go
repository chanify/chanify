package core

import (
	"strconv"
	"strings"

	"github.com/chanify/chanify/crypto"
	"github.com/chanify/chanify/model"
	"github.com/gin-gonic/gin"
)

func VerifyUser(ctx *gin.Context, key string) bool {
	sign, err := base64Encode.DecodeString(ctx.GetHeader("CHUserSign"))
	if err != nil {
		return false
	}
	data, _ := ctx.Get(gin.BodyBytesKey)
	return VerifySign(key, sign, data.([]byte))
}

func VerifyDevice(ctx *gin.Context, key string) bool {
	sign, err := base64Encode.DecodeString(ctx.GetHeader("CHDevSign"))
	if err != nil {
		return false
	}
	data, _ := ctx.Get(gin.BodyBytesKey)
	return VerifySign(key, sign, data.([]byte))
}

func VerifySign(key string, sign []byte, data []byte) bool {
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

func (c *Core) getToken(ctx *gin.Context) (*model.Token, error) {
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
	return c.parseToken(token)
}

func (c *Core) parseToken(token string) (*model.Token, error) {
	tk, err := model.ParseToken(token)
	if err != nil {
		return nil, err
	}
	if !c.logic.VerifyToken(tk) {
		return nil, model.ErrInvalidToken
	}
	return tk, nil
}

func parsePriority(priority string) int {
	if len(priority) > 0 {
		if p, err := strconv.Atoi(priority); err == nil {
			return p
		}
	}
	return 0
}

func parseImageContentType(data []byte) string {
	if string(data[:len(pngHeader)]) == pngHeader {
		return "image/png"
	} else {
		return "image/jpeg"
	}
}

type JsonString string

func (s *JsonString) UnmarshalJSON(data []byte) error {
	asString := strings.Trim(string(data), "\"")
	switch asString {
	case "1", "true", "TRUE", "True", "On", "on":
		*s = "1"
	case "0", "false", "FALSE", "False", "Off", "off", "none", "NONE", "null", "NULL":
		*s = ""
	default:
		*s = JsonString(asString)
	}
	return nil
}
