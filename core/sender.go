package core

import (
	"bytes"
	"net/http"
	"strings"
	"time"

	"github.com/chanify/chanify/logic"
	"github.com/chanify/chanify/model"
	"github.com/gin-gonic/gin"
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
	msg := model.NewMessage(token)
	msg, err = c.makeTextContent(msg, text, ctx.Query("title"), ctx.Query("copy"), ctx.Query("autocopy"), ctx.QueryArray("action"))
	if err != nil {
		ctx.JSON(http.StatusRequestEntityTooLarge, gin.H{"res": http.StatusRequestEntityTooLarge, "msg": "too large text content"})
		return
	}
	c.sendMsg(ctx, token, msg.SoundName(ctx.Query("sound")).SetPriority(parsePriority(ctx.Query("priority"))))
}

func (c *Core) handleUserSender(ctx *gin.Context) {
	uid, err := c.getUid(ctx)
	if len(uid) == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"res": http.StatusUnauthorized, "msg": "invalid user id"})
		return
	}
	text := ctx.Param("msg")
	if len(text) <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusNoContent, "msg": "no message content"})
		return
	}
	msg := &model.Message{}
	msg, err = c.makeTextContent(msg, text, ctx.Query("title"), ctx.Query("copy"), ctx.Query("autocopy"), ctx.QueryArray("action"))
	if err != nil {
		ctx.JSON(http.StatusRequestEntityTooLarge, gin.H{"res": http.StatusRequestEntityTooLarge, "msg": "too large text content"})
		return
	}
	c.sendUidMsg(ctx, uid, msg.SoundName(ctx.Query("sound")).SetPriority(parsePriority(ctx.Query("priority"))))
}

func (c *Core) handlePostSender(ctx *gin.Context) {
	params := &MsgParam{}
	params.Token, _ = c.getToken(ctx)
	params.Link = ctx.Query("link")
	params.Title = ctx.Query("title")
	params.Sound = ctx.Query("sound")
	params.AutoCopy = ctx.Query("autocopy")
	params.CopyText = ctx.Query("copy")
	params.Priority = parsePriority(ctx.Query("priority"))

	var err error
	var msg *model.Message = nil
	var parser func(c *Core, ctx *gin.Context) (*model.Message, error) = nil
	switch ctx.ContentType() {
	case "text/plain":
		params.ParsePlainText(ctx)
	case "application/json":
		params.ParseJSON(c, ctx)
	case "multipart/form-data":
		parser = params.ParseFormData
	case "image/png", "image/jpeg":
		parser = params.ParseImage
	case "audio/mpeg":
		parser = params.ParseAudio
	default:
		params.ParseForm(c, ctx)
	}
	if parser != nil {
		msg, err = parser(c, ctx)
		if err != nil {
			return
		}
	}
	if params.Token == nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"res": http.StatusUnauthorized, "msg": "invalid token format"})
		return
	}
	if msg == nil && len(params.Link) > 0 {
		msg = model.NewMessage(params.Token).LinkContent(params.Link)
	}
	if msg == nil {
		if len(params.Text) <= 0 {
			ctx.JSON(http.StatusNoContent, gin.H{"res": http.StatusNoContent, "msg": "no message content"})
			return
		}
		var err error
		msg, err = c.makeTextContent(model.NewMessage(params.Token), params.Text, params.Title, params.CopyText, params.AutoCopy, params.Actions)
		if err != nil {
			ctx.JSON(http.StatusRequestEntityTooLarge, gin.H{"res": http.StatusRequestEntityTooLarge, "msg": "too large text content"})
			return
		}
	}
	c.sendMsg(ctx, params.Token, msg.SoundName(params.Sound).SetPriority(params.Priority))
}

func (c *Core) sendUidDirect(ctx *gin.Context, uid string, msg *model.Message) {
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
	uuid, n := c.logic.SendAPNS(uid, out, devs, int(msg.Priority))
	if n <= 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"res": http.StatusNotFound, "msg": "no devices send success"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"request-uid": uuid})
}

func (c *Core) sendDirect(ctx *gin.Context, token *model.Token, msg *model.Message) {
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
	uuid, n := c.logic.SendAPNS(uid, out, devs, int(msg.Priority))
	if n <= 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"res": http.StatusNotFound, "msg": "no devices send success"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"request-uid": uuid})
}

func (c *Core) sendForward(ctx *gin.Context, token *model.Token, msg *model.Message) {
	msg.DisableToken()
	key, err := c.logic.GetUserKey(token.GetUserID())
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusBadRequest, "msg": "invalid user"})
		return
	}
	msg.EncryptContent(key)
	resp, err := http.Post(logic.APIEndpoint+"/rest/v1/push?token="+token.RawToken(), "application/x-protobuf", bytes.NewReader(msg.Marshal()))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"res": http.StatusInternalServerError, "msg": "send message failed"})
		return
	}
	reader := resp.Body
	defer reader.Close()
	ctx.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), reader, map[string]string{})
}

func (c *Core) sendUidMsg(ctx *gin.Context, uid string, msg *model.Message) {
	u, err := c.logic.GetUser(uid)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusBadRequest, "msg": "invalid user"})
		return
	}
	if u.IsServerless() {
		return
	}
	c.sendUidDirect(ctx, uid, msg)
}

func (c *Core) sendMsg(ctx *gin.Context, token *model.Token, msg *model.Message) {
	u, err := c.logic.GetUser(token.GetUserID())
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusBadRequest, "msg": "invalid user"})
		return
	}
	if u.IsServerless() {
		c.sendForward(ctx, token, msg)
		return
	}
	c.sendDirect(ctx, token, msg)
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
	return model.NewMessage(token).ImageContent(path, createThumbnail(data), len(data)), nil
}

func (c *Core) saveUploadAudio(ctx *gin.Context, token *model.Token, data []byte) (*model.Message, error) {
	if len(data) <= 0 {
		ctx.JSON(http.StatusNoContent, gin.H{"res": http.StatusNoContent, "msg": "no audio content"})
		return nil, ErrNoContent
	}
	path, err := c.logic.SaveFile("audios", data)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": http.StatusBadRequest, "msg": "invalid audio content"})
		return nil, ErrInvalidContent
	}
	return model.NewMessage(token).AudioContent(path, 0, len(data)), nil
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

func (c *Core) makeTextContent(msg *model.Message, text string, title string, copytext string, autocopy string, actions []string) (*model.Message, error) {
	if len(actions) > 0 {
		return c.makeActionContent(msg, text, title, actions)
	}
	if len(copytext) > 1000 {
		return nil, ErrTooLargeContent
	}
	if len(text)+len(title) < 1200 {
		return msg.TextContent(text, title, copytext, autocopy), nil
	}
	if !c.logic.CanFileStore() {
		return nil, ErrTooLargeContent
	}
	txts := []string{}
	if len(title) > 0 {
		txts = append(txts, title)
	}
	if len(text) > 0 {
		txts = append(txts, text)
	}
	data := []byte(strings.Join(txts, "\n\n"))
	path, err := c.logic.SaveFile("files", data)
	if err != nil {
		return nil, err
	}
	if len(title) > 100 {
		title = string([]rune(title)[:100]) + "⋯"
	}
	return msg.TextFileContent(path, "text.txt", title, string([]rune(text)[:100])+"⋯", len(data)), nil
}

func (c *Core) makeActionContent(msg *model.Message, text string, title string, actions []string) (*model.Message, error) {
	if len(actions) > 4 {
		actions = actions[:4]
	}
	l := len(title) + len(text)
	for _, act := range actions {
		l += len(act)
	}
	if l > 2000 {
		return nil, ErrTooLargeContent
	}
	return msg.ActionContent(text, title, actions), nil
}
