package core

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/chanify/chanify/logic"
	"github.com/chanify/chanify/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (c *Core) handleSender(ctx *gin.Context) {
	token, err := c.getToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"res": http.StatusUnauthorized, "msg": "invalid token"})
		return
	}
	c.sendMsg(ctx, token, ctx.Param("msg"))
}

func (c *Core) handlePostSender(ctx *gin.Context) {
	token, _ := c.getToken(ctx)
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

func (c *Core) SendDirect(ctx *gin.Context, token *model.Token, msg *model.Message) {
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
	out := msg.EncryptData(key, uint64(time.Now().UTC().UnixNano()))
	c.logic.SendAPNS(uid, msg, out, devs) // nolint: errcheck
	ctx.JSON(http.StatusOK, gin.H{"request-uid": uuid.New().String()})
}

func (c *Core) SendForward(ctx *gin.Context, token *model.Token, msg *model.Message) {
	msg.DisableToken()
	key, err := c.logic.GetUserKey(token.GetUserID())
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusBadRequest, "msg": "invalid user"})
		return
	}
	msg.EncryptContent(key)
	resp, err := http.Post(logic.ApiEndpoint+"/rest/v1/push?token="+token.RawToken(), "application/x-protobuf", bytes.NewReader(msg.Marshal()))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"res": http.StatusInternalServerError, "msg": "send message failed"})
		return
	}
	reader := resp.Body
	defer reader.Close()
	ctx.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), reader, map[string]string{})
}

func (c *Core) sendMsg(ctx *gin.Context, token *model.Token, text string) {
	if token == nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"res": http.StatusUnauthorized, "msg": "invalid token format"})
		return
	}
	if len(text) <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusNoContent, "msg": "no message content"})
		return
	}
	u, err := c.logic.GetUser(token.GetUserID())
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusBadRequest, "msg": "invalid user"})
		return
	}
	msg := model.NewMessage(token).TextContent(text)
	if u.IsServerless() {
		c.SendForward(ctx, token, msg)
		return
	}
	c.SendDirect(ctx, token, msg)
}
