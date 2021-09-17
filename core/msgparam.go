package core

import (
	"io/ioutil"
	"mime/multipart"
	"strconv"
	"strings"
	"time"

	"github.com/chanify/chanify/model"
	"github.com/gin-gonic/gin"
)

// TimeContent define timeline content
type TimeContent struct {
	Code      string
	Timestamp *time.Time
	Items     []*model.MsgTimeItem
}

// MsgParam parse message parameters
type MsgParam struct {
	Token       *model.Token
	Text        string
	Link        string
	Title       string
	Sound       string
	AutoCopy    string
	CopyText    string
	Priority    int
	Actions     []string
	TimeContent TimeContent
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
		Timeline struct {
			Code     string                 `json:"code"`
			Timstamp interface{}            `json:"timestamp,omitempty"`
			Items    map[string]interface{} `json:"items"`
		} `json:"timeline,omitempty"`
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
		if len(m.TimeContent.Code) <= 0 {
			m.TimeContent.Code = params.Timeline.Code
			m.TimeContent.Timestamp = parseTimestamp(params.Timeline.Timstamp)
			m.TimeContent.Items = parseTimeContentItems(params.Timeline.Items)
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
	if len(m.TimeContent.Code) <= 0 {
		m.TimeContent.Code = ctx.PostForm("timeline-code")
		m.TimeContent.Timestamp = parseTimestamp(ctx.PostForm("timeline-timestamp"))
		m.TimeContent.Items = parseTimeContentStringItems(ctx.PostFormMap("timeline-items"))
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
		m.TimeContent.Code = tryFormValue(form, "timeline-code", m.TimeContent.Code)
		if len(m.TimeContent.Code) > 0 {
			m.TimeContent.Timestamp = tryFormTimestamp(form, "timeline-timestamp", m.TimeContent.Timestamp)
			m.TimeContent.Items = tryFormMap(form, "timeline-items", m.TimeContent.Items)
		}
		if m.Token != nil && c.logic.CanFileStore() {
			if data, _, err := readFileFromForm(form, "image"); err == nil {
				msg, err = c.saveUploadImage(ctx, m.Token, data)
				if err != nil {
					return nil, err
				}
			}
			if data, fname, err := readFileFromForm(form, "audio"); err == nil {
				title := m.Title
				if len(title) <= 0 {
					title = fileBaseName(fname)
				}
				msg, err = c.saveUploadAudio(ctx, m.Token, title, data)
				if err != nil {
					return nil, err
				}
			}
			if data, fname, err := readFileFromForm(form, "file"); err == nil {
				msg, err = c.saveUploadFile(ctx, m.Token, data, fname, m.Text, m.Actions)
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
		msg, err = c.saveUploadAudio(ctx, m.Token, m.Title, data)
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

func parseTimeContentItems(items map[string]interface{}) []*model.MsgTimeItem {
	lst := []*model.MsgTimeItem{}
	for k, v := range items {
		switch val := v.(type) {
		case int:
			lst = append(lst, &model.MsgTimeItem{
				Name:  k,
				Value: int64(val),
			})
		case int64:
			lst = append(lst, &model.MsgTimeItem{
				Name:  k,
				Value: val,
			})
		case float32:
			lst = append(lst, &model.MsgTimeItem{
				Name:  k,
				Value: float64(val),
			})
		case float64:
			lst = append(lst, &model.MsgTimeItem{
				Name:  k,
				Value: val,
			})
		case string:
			if strings.ContainsRune(val, '.') {
				if vv, err := strconv.ParseFloat(val, 64); err == nil {
					lst = append(lst, &model.MsgTimeItem{
						Name:  k,
						Value: vv,
					})
					continue
				}
			} else {
				if vv, err := strconv.ParseInt(val, 10, 64); err == nil {
					lst = append(lst, &model.MsgTimeItem{
						Name:  k,
						Value: vv,
					})
					continue
				}
			}
			lst = append(lst, &model.MsgTimeItem{
				Name:  k,
				Value: 0,
			})
		default:
			lst = append(lst, &model.MsgTimeItem{
				Name:  k,
				Value: 0,
			})
		}
	}
	return lst
}

func parseTimeContentStringItems(items map[string]string) []*model.MsgTimeItem {
	lst := []*model.MsgTimeItem{}
	for k, v := range items {
		if strings.ContainsRune(v, '.') {
			if vv, err := strconv.ParseFloat(v, 64); err == nil {
				lst = append(lst, &model.MsgTimeItem{
					Name:  k,
					Value: vv,
				})
				continue
			}
		} else {
			if vv, err := strconv.ParseInt(v, 10, 64); err == nil {
				lst = append(lst, &model.MsgTimeItem{
					Name:  k,
					Value: vv,
				})
				continue
			}
		}
		lst = append(lst, &model.MsgTimeItem{
			Name:  k,
			Value: 0,
		})
	}
	return lst
}

func parseTimestamp(t interface{}) *time.Time {
	switch val := t.(type) {
	case int:
		v := time.Unix(int64(val)/1000, int64(val)%1000*1e6)
		return &v
	case uint:
		v := time.Unix(int64(val)/1000, int64(val)%1000*1e6)
		return &v
	case int64:
		v := time.Unix(val/1000, val%1000*1e6)
		return &v
	case uint64:
		v := time.Unix(int64(val)/1000, int64(val)%1000*1e6)
		return &v
	case string:
		if v, err := time.Parse(time.RFC3339Nano, val); err == nil {
			return &v
		} else if v, err := strconv.ParseUint(val, 10, 64); err == nil {
			vv := time.Unix(int64(v)/1000, int64(v)%1000*1e6)
			return &vv
		}
	}
	return nil
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

func tryFormMap(form *multipart.Form, name string, items []*model.MsgTimeItem) []*model.MsgTimeItem {
	if len(items) <= 0 {
		l := len(name)
		values := map[string]string{}
		for k, v := range form.Value {
			if strings.HasPrefix(k, name) && len(k) > l+2 && k[l] == '[' && k[len(k)-1] == ']' && len(v) > 0 {
				values[strings.TrimSpace(k[l+1:len(k)-1])] = v[0]
			}
		}
		return parseTimeContentStringItems(values)
	}
	return items
}

func tryFormTimestamp(form *multipart.Form, name string, ts *time.Time) *time.Time {
	if ts == nil {
		ts = parseTimestamp(form.Value[name])
	}
	return ts
}
