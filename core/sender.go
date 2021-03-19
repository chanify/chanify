package core

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/chanify/chanify/logic"
	"github.com/chanify/chanify/model"
	"github.com/chanify/chanify/pb"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

func (c *Core) handleSender(ctx *gin.Context) {
	token, _ := getToken(ctx)
	c.sendMsg(ctx, token, ctx.Param("msg"))
}

func (c *Core) handlePostSender(ctx *gin.Context) {
	token, _ := getToken(ctx)
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
			if token == nil {
				tks := form.Value["token"]
				if len(tks) > 0 {
					token, _ = model.ParseToken(tks[0])
				}
			}
		}
	default:
		msg = ctx.PostForm("text")
	}
	c.sendMsg(ctx, token, msg)
}

func (c *Core) SendDirect(ctx *gin.Context, token *model.Token, text string) {
	uid := token.GetUserID()
	key, err := c.logic.GetUserKey(uid)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusBadRequest, "msg": "invalid user"})
		return
	}
	devs, err := c.logic.GetDevices(uid)
	if err != nil || len(devs) <= 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"res": http.StatusNotFound, "msg": "no devices found"})
		return
	}
	msg := c.logic.CreateMessage(token)
	msg.Content, _ = proto.Marshal(&pb.MsgContent{
		Type: pb.MsgType_Text,
		Text: text,
	})
	data, _ := proto.Marshal(msg)
	aesgcm, _ := NewAESGCM(key)
	nonce := make([]byte, 12)
	nonce[0] = 0x01
	nonce[1] = 0x01
	nonce[2] = 0x00
	nonce[3] = 0x08
	binary.BigEndian.PutUint64(nonce[4:], uint64(time.Now().UTC().UnixNano()))

	tag := key[32 : 32+32]
	out := aesgcm.Seal(nil, nonce, data, tag)
	out = append(nonce, out...)
	c.logic.SendAPNS(uid, out, devs) // nolint: errcheck
	ctx.JSON(http.StatusOK, gin.H{"request-uid": uuid.New().String()})
}

func (c *Core) SendForward(ctx *gin.Context, token *model.Token, msg string) {
	content, _ := proto.Marshal(&pb.MsgContent{
		Type: pb.MsgType_Text,
		Text: msg,
	})
	key, err := c.logic.GetUserKey(token.GetUserID())
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusBadRequest, "msg": "invalid user"})
		return
	}
	aesgcm, _ := NewAESGCM(key) // nolint: errcheck
	nonce := make([]byte, 12)
	rand.Read(nonce) // nolint: errcheck
	data := aesgcm.Seal(nil, nonce, content, key[32:32+32])
	data = append(nonce, data...)
	m, _ := proto.Marshal(&pb.Message{Ciphertext: data})
	resp, err := http.Post(logic.ApiEndpoint+"/rest/v1/push?token="+token.RawToken(), "application/x-protobuf", bytes.NewReader(m))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"res": http.StatusInternalServerError, "msg": "send message failed"})
		return
	}
	reader := resp.Body
	defer reader.Close()
	ctx.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), reader, map[string]string{})
}

func (c *Core) sendMsg(ctx *gin.Context, token *model.Token, msg string) {
	if token == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusBadRequest, "msg": "invalid token format"})
		return
	}
	if len(msg) <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusNoContent, "msg": "no message"})
		return
	}
	u, err := c.logic.GetUser(token.GetUserID())
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusBadRequest, "msg": "invalid user"})
		return
	}
	if u.IsServerless() {
		c.SendForward(ctx, token, msg)
		return
	}
	c.SendDirect(ctx, token, msg)
}
