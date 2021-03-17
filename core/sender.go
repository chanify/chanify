package core

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io/ioutil"
	"net/http"

	"github.com/chanify/chanify/pb"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

func (c *Core) handleSender(ctx *gin.Context) {
	token := ctx.GetHeader("token")
	if len(token) <= 0 {
		token = ctx.Query("token")
		if len(token) <= 0 {
			token = ctx.Param("token")
		}
	}
	var msg string
	switch ctx.ContentType() {
	case "text/plain":
		defer ctx.Request.Body.Close()
		if d, err := ioutil.ReadAll(ctx.Request.Body); err == nil {
			msg = string(d)
		}
	case "multipart/form-data":
		if form, err := ctx.MultipartForm(); err == nil {
			ts := form.Value["text"]
			if len(ts) > 0 {
				msg = ts[0]
			}
			if len(token) <= 0 {
				tks := form.Value["token"]
				if len(tks) > 0 {
					token = tks[0]
				}
			}
		}
	default:
		msg = ctx.PostForm("text")
	}
	tk, err := NewToken(token)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusBadRequest, "msg": "invalid token"})
		return
	}
	c.sendMsg(ctx, tk, msg)
}

func (c *Core) sendMsg(ctx *gin.Context, token *Token, msg string) {
	uuid := uuid.New().String()
	content, err := proto.Marshal(&pb.MsgContent{
		Type: pb.MsgType_Text,
		Text: msg,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"res": http.StatusInternalServerError, "msg": "format message content failed"})
		return
	}
	key, err := c.logic.GetUserKey(token.UserId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusBadRequest, "msg": "invalid user"})
		return
	}
	block, err := aes.NewCipher(key[:32])
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"res": http.StatusInternalServerError, "msg": "unknown error"})
		return
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"res": http.StatusInternalServerError, "msg": "unknown error"})
		return

	}
	nonce := make([]byte, 12)
	rand.Read(nonce) // nolint: errcheck
	data := aesgcm.Seal(nil, nonce, content, key[32:32+32])
	data = append(nonce, data...)
	m, err := proto.Marshal(&pb.Message{Ciphertext: data})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"res": http.StatusInternalServerError, "msg": "format message failed"})
		return
	}
	resp, err := http.Post("https://api.chanify.net/rest/v1/push?token="+token.String(), "application/x-protobuf", bytes.NewReader(m))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"res": http.StatusInternalServerError, "msg": "send message failed"})
		return
	}
	if resp.StatusCode != http.StatusOK {
		ctx.JSON(http.StatusInternalServerError, gin.H{"res": http.StatusInternalServerError, "msg": "send push message failed"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"request-uid": uuid})
}
