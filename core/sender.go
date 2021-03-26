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
	text := ctx.Param("msg")
	if len(text) <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusNoContent, "msg": "no message content"})
		return
	}
	msg := model.NewMessage(token).TextContent(text, ctx.Query("title")).
		SoundName(ctx.Query("sound")).SetPriority(parsePriority(ctx.Query("priority")))
	c.sendMsg(ctx, token, msg)
}

func (c *Core) handlePostSender(ctx *gin.Context) {
	token, _ := c.getToken(ctx)
	var text string
	title := ctx.Query("title")
	sound := ctx.Query("sound")
	priority := parsePriority(ctx.Query("priority"))
	switch ctx.ContentType() {
	case "text/plain":
		defer ctx.Request.Body.Close()
		if d, err := ioutil.ReadAll(ctx.Request.Body); err == nil {
			text = string(d)
		}
	case "application/json":
		defer ctx.Request.Body.Close()
		var params struct {
			Token    string     `json:"token,omitempty"`
			Title    string     `json:"title,omitempty"`
			Text     string     `json:"text,omitempty"`
			Sound    JsonString `json:"sound,omitempty"`
			Priority int        `json:"priority,omitempty"`
		}
		if err := ctx.BindJSON(&params); err == nil {
			if token == nil && len(params.Token) > 0 {
				token, _ = model.ParseToken(params.Token)
			}
			if len(title) <= 0 && len(params.Title) > 0 {
				title = params.Title
			}
			if len(sound) <= 0 && len(params.Sound) > 0 {
				sound = string(params.Sound)
			}
			if priority <= 0 {
				priority = params.Priority
			}
			text = params.Text
		}
	case "multipart/form-data":
		if form, err := ctx.MultipartForm(); err == nil {
			ts := form.Value["text"]
			if len(ts) > 0 {
				text = ts[0]
			}
			if token == nil {
				tks := form.Value["token"]
				if len(tks) > 0 {
					token, _ = model.ParseToken(tks[0])
				}
			}
			if len(title) <= 0 {
				ts := form.Value["title"]
				if len(ts) > 0 {
					title = ts[0]
				}
			}
			if len(sound) <= 0 {
				ss := form.Value["sound"]
				if len(ss) > 0 {
					sound = ss[0]
				}
			}
			if priority <= 0 {
				ps := form.Value["priority"]
				if len(ps) > 0 {
					priority = parsePriority(ps[0])
				}
			}
		}
	default:
		text = ctx.PostForm("text")
		if token == nil {
			token, _ = model.ParseToken(ctx.PostForm("token"))
		}
		if len(sound) <= 0 {
			sound = ctx.PostForm("sound")
		}
		if priority <= 0 {
			priority = parsePriority(ctx.PostForm("priority"))
		}
	}
	if token == nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"res": http.StatusUnauthorized, "msg": "invalid token format"})
		return
	}
	if len(text) <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusNoContent, "msg": "no message content"})
		return
	}
	msg := model.NewMessage(token).TextContent(text, title).SoundName(sound).SetPriority(priority)
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
	c.logic.SendAPNS(uid, msg, out, devs, int(msg.Priority)) // nolint: errcheck
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

func (c *Core) sendMsg(ctx *gin.Context, token *model.Token, msg *model.Message) {
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
