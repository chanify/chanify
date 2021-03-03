package core

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/chanify/chanify/pb"
	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
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
	sendMsg(ctx, token, msg)
}

func sendMsg(ctx *gin.Context, token string, msg string) {
	if len(token) <= 0 || len(msg) <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusBadRequest, "msg": "bad request"})
		return
	}
	uuid := uuid.New().String()
	content, err := proto.Marshal(&pb.MsgContent{
		Type: pb.MsgType_Text,
		Text: msg,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"res": http.StatusInternalServerError, "msg": "format message content failed"})
		return
	}
	m, err := proto.Marshal(&pb.Message{
		Content: content,
		//Ciphertext: content,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"res": http.StatusInternalServerError, "msg": "format message failed"})
		return
	}
	resp, err := http.Post("https://api.chanify.net/rest/v1/push?token="+token, "application/x-protobuf", bytes.NewReader(m))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"res": http.StatusInternalServerError, "msg": "send message failed"})
		return
	}
	if resp.StatusCode != http.StatusOK {
		log.Println("4444444", resp.StatusCode)
		ctx.JSON(http.StatusInternalServerError, gin.H{"res": http.StatusInternalServerError, "msg": "send push message failed"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"request-uid": uuid})
}
