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
	msg := model.NewMessage(token).TextContent(text, ctx.Query("title"), ctx.Query("copy"), ctx.Query("autocopy")).
		SoundName(ctx.Query("sound")).SetPriority(parsePriority(ctx.Query("priority")))
	c.sendMsg(ctx, token, msg)
}

func (c *Core) handlePostSender(ctx *gin.Context) {
	token, _ := c.getToken(ctx)
	var text string
	link := ctx.Query("link")
	title := ctx.Query("title")
	sound := ctx.Query("sound")
	autocopy := ctx.Query("autocopy")
	copytext := ctx.Query("copy")
	priority := parsePriority(ctx.Query("priority"))
	var msg *model.Message = nil
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
			Copy     string     `json:"copy,omitempty"`
			AutoCopy JsonString `json:"autocopy,omitempty"`
			Link     string     `json:"link,omitempty"`
			Sound    JsonString `json:"sound,omitempty"`
			Priority int        `json:"priority,omitempty"`
		}
		if err := ctx.BindJSON(&params); err == nil {
			if token == nil && len(params.Token) > 0 {
				token, _ = c.parseToken(params.Token)
			}
			if len(link) <= 0 && len(params.Link) > 0 {
				link = params.Link
			}
			if len(title) <= 0 && len(params.Title) > 0 {
				title = params.Title
			}
			if len(sound) <= 0 && len(params.Sound) > 0 {
				sound = string(params.Sound)
			}
			if len(copytext) <= 0 && len(params.Copy) > 0 {
				copytext = params.Copy
			}
			if len(autocopy) <= 0 && len(params.AutoCopy) > 0 {
				autocopy = string(params.AutoCopy)
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
					token, _ = c.parseToken(tks[0])
				}
			}
			if len(title) <= 0 {
				ts := form.Value["title"]
				if len(ts) > 0 {
					title = ts[0]
				}
			}
			if len(link) <= 0 {
				ls := form.Value["link"]
				if len(ls) > 0 {
					link = ls[0]
				}
			}
			if len(copytext) <= 0 {
				cs := form.Value["copy"]
				if len(cs) > 0 {
					copytext = cs[0]
				}
			}
			if len(autocopy) <= 0 {
				as := form.Value["autocopy"]
				if len(as) > 0 {
					autocopy = as[0]
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
			if token != nil && c.logic.CanFileStore() {
				fs := form.File["image"]
				if len(fs) > 0 {
					if fp, err := fs[0].Open(); err == nil {
						defer fp.Close()
						data, _ := ioutil.ReadAll(fp)
						msg, err = c.saveUploadImage(ctx, token, data)
						if err != nil {
							return
						}
					}
				}
				fs = form.File["file"]
				if len(fs) > 0 {
					if fp, err := fs[0].Open(); err == nil {
						defer fp.Close()
						data, _ := ioutil.ReadAll(fp)
						msg, err = c.saveUploadFile(ctx, token, data, fs[0].Filename, text)
						if err != nil {
							return
						}
					}
				}
			}
		}
	case "image/png", "image/jpeg":
		if token != nil && c.logic.CanFileStore() {
			var err error
			data, _ := ctx.GetRawData()
			msg, err = c.saveUploadImage(ctx, token, data)
			if err != nil {
				return
			}
		}
	default:
		text = ctx.PostForm("text")
		if token == nil {
			token, _ = c.parseToken(ctx.PostForm("token"))
		}
		if len(link) <= 0 {
			link = ctx.PostForm("link")
		}
		if len(copytext) <= 0 {
			copytext = ctx.PostForm("copy")
		}
		if len(autocopy) <= 0 {
			autocopy = ctx.PostForm("autocopy")
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
	if msg == nil && len(link) > 0 {
		msg = model.NewMessage(token).LinkContent(link)
	}
	if msg == nil {
		if len(text) <= 0 {
			ctx.JSON(http.StatusNoContent, gin.H{"res": http.StatusNoContent, "msg": "no message content"})
			return
		}
		msg = model.NewMessage(token).TextContent(text, title, copytext, autocopy)
	}
	c.sendMsg(ctx, token, msg.SoundName(sound).SetPriority(priority))
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
	if len(out) > 4000 {
		ctx.JSON(http.StatusRequestEntityTooLarge, gin.H{"res": http.StatusRequestEntityTooLarge, "msg": "message body too large"})
		return
	}
	if n := c.logic.SendAPNS(uid, out, devs, int(msg.Priority)); n <= 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"res": http.StatusNotFound, "msg": "no devices send success"})
		return
	}
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

func (c *Core) saveUploadImage(ctx *gin.Context, token *model.Token, data []byte) (*model.Message, error) {
	if len(data) <= 0 {
		ctx.JSON(http.StatusNoContent, gin.H{"res": http.StatusNoContent, "msg": "no image content"})
		return nil, ErrNoContent
	}
	path, err := c.logic.SaveFile("images", data)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusBadRequest, "msg": "invalid image content"})
		return nil, ErrInvalidContent
	}
	return model.NewMessage(token).ImageContent(path, CreateThumbnail(data), len(data)), nil
}

func (c *Core) saveUploadFile(ctx *gin.Context, token *model.Token, data []byte, filename string, desc string) (*model.Message, error) {
	if len(data) <= 0 {
		ctx.JSON(http.StatusNoContent, gin.H{"res": http.StatusNoContent, "msg": "no file content"})
		return nil, ErrNoContent
	}
	path, err := c.logic.SaveFile("files", data)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusBadRequest, "msg": "invalid file content"})
		return nil, ErrInvalidContent
	}
	return model.NewMessage(token).FileContent(path, filename, desc, len(data)), nil
}
