package core

import (
	"io/ioutil"
	"mime/multipart"

	"github.com/chanify/chanify/model"
	"github.com/gin-gonic/gin"
)

// MsgParam parse message parameters
type MsgParam struct {
	Token    *model.Token
	Text     string
	Link     string
	Title    string
	Sound    string
	AutoCopy string
	CopyText string
	Priority int
	Actions  []string
}

// ParsePlainText process text/plain
func (m *MsgParam) ParsePlainText(ctx *gin.Context) {
	defer ctx.Request.Body.Close()
	if d, err := ioutil.ReadAll(ctx.Request.Body); err == nil {
		m.Text = string(d)
	}
}

// ParseJSON process application/json
func (m *MsgParam) ParseJSON(c *Core, ctx *gin.Context) {
	defer ctx.Request.Body.Close()
	var params struct {
		Token    string     `json:"token,omitempty"`
		Title    string     `json:"title,omitempty"`
		Text     string     `json:"text,omitempty"`
		Copy     string     `json:"copy,omitempty"`
		AutoCopy JSONString `json:"autocopy,omitempty"`
		Link     string     `json:"link,omitempty"`
		Sound    JSONString `json:"sound,omitempty"`
		Priority int        `json:"priority,omitempty"`
		Actions  []string   `json:"actions,omitempty"`
	}
	if err := ctx.BindJSON(&params); err == nil {
		if m.Token == nil && len(params.Token) > 0 {
			m.Token, _ = c.parseToken(params.Token)
		}
		m.Link = tryStringValue(m.Link, params.Link)
		m.Title = tryStringValue(m.Title, params.Title)
		if len(m.Sound) <= 0 && len(params.Sound) > 0 {
			m.Sound = string(params.Sound)
		}
		m.CopyText = tryStringValue(m.CopyText, params.Copy)
		if len(m.AutoCopy) <= 0 && len(params.AutoCopy) > 0 {
			m.AutoCopy = string(params.AutoCopy)
		}
		if len(m.Actions) <= 0 {
			m.Actions = params.Actions
		}
		if m.Priority <= 0 {
			m.Priority = params.Priority
		}
		m.Text = params.Text
	}
}

// ParseForm process form
func (m *MsgParam) ParseForm(c *Core, ctx *gin.Context) {
	m.Text = ctx.PostForm("text")
	if m.Token == nil {
		m.Token, _ = c.parseToken(ctx.PostForm("token"))
	}
	if len(m.Link) <= 0 {
		m.Link = ctx.PostForm("link")
	}
	if len(m.CopyText) <= 0 {
		m.CopyText = ctx.PostForm("copy")
	}
	if len(m.AutoCopy) <= 0 {
		m.AutoCopy = ctx.PostForm("autocopy")
	}
	if len(m.Sound) <= 0 {
		m.Sound = ctx.PostForm("sound")
	}
	if len(m.Actions) <= 0 {
		m.Actions = ctx.PostFormArray("action")
	}
	if m.Priority <= 0 {
		m.Priority = parsePriority(ctx.PostForm("priority"))
	}
}

// ParseFormData process multipart/form-data
func (m *MsgParam) ParseFormData(c *Core, ctx *gin.Context) (*model.Message, error) {
	var msg *model.Message = nil
	if form, err := ctx.MultipartForm(); err == nil {
		ts := form.Value["text"]
		if len(ts) > 0 {
			m.Text = ts[0]
		}
		if m.Token == nil {
			tks := form.Value["token"]
			if len(tks) > 0 {
				m.Token, _ = c.parseToken(tks[0])
			}
		}
		m.Title = tryFormValue(form, "title", m.Title)
		m.Link = tryFormValue(form, "link", m.Link)
		m.CopyText = tryFormValue(form, "copy", m.CopyText)
		m.AutoCopy = tryFormValue(form, "autocopy", m.AutoCopy)
		m.Sound = tryFormValue(form, "sound", m.Sound)
		m.Actions = tryFormValues(form, "action", m.Actions)
		m.parsePriorityFromForm(form)
		if m.Token != nil && c.logic.CanFileStore() {
			if data, _, err := readFileFromForm(form, "image"); err == nil {
				msg, err = c.saveUploadImage(ctx, m.Token, data)
				if err != nil {
					return nil, err
				}
			}
			if data, _, err := readFileFromForm(form, "audio"); err == nil {
				msg, err = c.saveUploadAudio(ctx, m.Token, data)
				if err != nil {
					return nil, err
				}
			}
			if data, fname, err := readFileFromForm(form, "file"); err == nil {
				msg, err = c.saveUploadFile(ctx, m.Token, data, fname, m.Text)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return msg, nil
}

// ParseImage process image
func (m *MsgParam) ParseImage(c *Core, ctx *gin.Context) (*model.Message, error) {
	var msg *model.Message = nil
	if m.Token != nil && c.logic.CanFileStore() {
		var err error
		data, _ := ctx.GetRawData()
		msg, err = c.saveUploadImage(ctx, m.Token, data)
		if err != nil {
			return nil, err
		}
	}
	return msg, nil
}

// ParseAudio process audio
func (m *MsgParam) ParseAudio(c *Core, ctx *gin.Context) (*model.Message, error) {
	var msg *model.Message = nil
	if m.Token != nil && c.logic.CanFileStore() {
		var err error
		data, _ := ctx.GetRawData()
		msg, err = c.saveUploadAudio(ctx, m.Token, data)
		if err != nil {
			return nil, err
		}
	}
	return msg, nil
}

func (m *MsgParam) parsePriorityFromForm(form *multipart.Form) {
	if m.Priority <= 0 {
		ps := form.Value["priority"]
		if len(ps) > 0 {
			m.Priority = parsePriority(ps[0])
		}
	}
}

func readFileFromForm(form *multipart.Form, name string) ([]byte, string, error) {
	fs := form.File[name]
	if len(fs) > 0 {
		if fp, err := fs[0].Open(); err == nil {
			defer fp.Close()
			data, err := ioutil.ReadAll(fp)
			return data, fs[0].Filename, err
		}
	}
	return nil, "", ErrNoContent
}

func tryStringValue(value string, newValue string) string {
	if len(value) <= 0 && len(newValue) > 0 {
		value = newValue
	}
	return value
}

func tryFormValue(form *multipart.Form, name string, value string) string {
	if len(value) <= 0 {
		vs := form.Value[name]
		if len(vs) > 0 {
			return vs[0]
		}
	}
	return value
}

func tryFormValues(form *multipart.Form, name string, value []string) []string {
	if len(value) <= 0 {
		value = form.Value[name]
	}
	return value
}
