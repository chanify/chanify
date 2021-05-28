package model

import (
	"crypto/rand"
	"encoding/binary"
	"strings"

	"github.com/chanify/chanify/pb"
	"google.golang.org/protobuf/proto"
)

// Message for notification
type Message struct {
	pb.Message
}

// NewMessage with sender token
func NewMessage(tk *Token) *Message {
	m := &Message{}
	m.From = tk.GetNodeID()
	m.Channel = tk.GetChannel()
	return m
}

// DisableToken clear token
func (m *Message) DisableToken() *Message {
	m.From = nil
	m.Channel = nil
	return m
}

// LinkContent set link notification
func (m *Message) LinkContent(link string) *Message {
	ctx := &pb.MsgContent{
		Type: pb.MsgType_Link,
		Link: link,
	}
	m.Content, _ = proto.Marshal(ctx)
	return m
}

// TextContent set text notification
func (m *Message) TextContent(text string, title string, copytext string, autocopy string) *Message {
	ctx := &pb.MsgContent{
		Type: pb.MsgType_Text,
		Text: text,
	}
	if len(title) > 0 {
		ctx.Title = title
	}
	if len(copytext) > 0 {
		ctx.Copytext = copytext
	}
	if len(autocopy) > 0 {
		ctx.Flags = 1
	}
	m.Content, _ = proto.Marshal(ctx)
	return m
}

// ActionContent set custom action notification
func (m *Message) ActionContent(text string, title string, actions []string) *Message {
	ctx := &pb.MsgContent{
		Type: pb.MsgType_Action,
	}
	if len(title) > 0 {
		ctx.Title = title
	}
	if len(text) > 0 {
		ctx.Text = text
	}
	if len(actions) > 4 {
		actions = actions[:4]
	}
	ctx.Actions = []*pb.ActionItem{}
	for _, act := range actions {
		ss := strings.SplitN(act, "|", 2)
		if len(ss) > 1 {
			item := &pb.ActionItem{
				Type: pb.ActType_ActURL,
				Name: ss[0],
				Link: ss[1],
			}
			ctx.Actions = append(ctx.Actions, item)
		}
	}
	m.Content, _ = proto.Marshal(ctx)
	return m
}

// FileContent set file notification
func (m *Message) FileContent(path string, filename string, desc string, size int) *Message {
	ctx := &pb.MsgContent{
		Type:     pb.MsgType_File,
		File:     path,
		Filename: filename,
		Size:     uint64(size),
	}
	if len(desc) > 0 {
		ctx.Text = desc
	}
	m.Content, _ = proto.Marshal(ctx)
	return m
}

// ImageContent set image notification
func (m *Message) ImageContent(path string, t *Thumbnail, size int) *Message {
	ctx := &pb.MsgContent{
		Type: pb.MsgType_Image,
		File: path,
		Size: uint64(size),
	}
	if t != nil {
		ctx.Thumbnail = &pb.Thumbnail{
			Width:  int32(t.width),
			Height: int32(t.height),
			Data:   t.preview,
		}
	}
	m.Content, _ = proto.Marshal(ctx)
	return m
}

// AudioContent set audio notification
func (m *Message) AudioContent(path string, duration uint64, size int) *Message {
	ctx := &pb.MsgContent{
		Type: pb.MsgType_Audio,
		File: path,
		Size: uint64(size),
	}
	m.Content, _ = proto.Marshal(ctx)
	return m
}

// TextFileContent set text file notification
func (m *Message) TextFileContent(path string, filename string, title string, desc string, size int) *Message {
	ctx := &pb.MsgContent{
		Type:     pb.MsgType_File,
		File:     path,
		Filename: filename,
		Size:     uint64(size),
	}
	if len(title) > 0 {
		ctx.Title = title
	}
	if len(desc) > 0 {
		ctx.Text = desc
	}
	m.Content, _ = proto.Marshal(ctx)
	return m
}

// SoundName set notification sound
func (m *Message) SoundName(sound string) *Message {
	if len(sound) > 0 {
		m.Sound = &pb.Sound{Name: sound}
	}
	return m
}

// SetPriority set notification priority
func (m *Message) SetPriority(priority int) *Message {
	if priority > 0 && priority < 0x7fffffff {
		m.Priority = int32(priority)
	}
	return m
}

// EncryptContent return encrypted content with key
func (m *Message) EncryptContent(key []byte) {
	if m.Content != nil {
		aesgcm, _ := NewAESGCM(key)
		nonce := make([]byte, 12)
		rand.Read(nonce) // nolint: errcheck
		data := aesgcm.Seal(nil, nonce, m.Content, key[32:32+32])
		m.Ciphertext = append(nonce, data...)
		m.Content = nil
	}
}

// EncryptData return encrypted body with key & timestamp
func (m *Message) EncryptData(key []byte, ts uint64) []byte {
	aesgcm, _ := NewAESGCM(key)
	nonce := make([]byte, 12)
	nonce[0] = 0x01
	nonce[1] = 0x01
	nonce[2] = 0x00
	nonce[3] = 0x08
	binary.BigEndian.PutUint64(nonce[4:], ts)

	tag := key[32 : 32+32]
	out := aesgcm.Seal(nil, nonce, m.Marshal(), tag)
	return append(nonce, out...)
}

// Marshal return binary data
func (m *Message) Marshal() []byte {
	data, _ := proto.Marshal(&m.Message)
	return data
}
