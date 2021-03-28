package model

import (
	"crypto/rand"
	"encoding/binary"

	"github.com/chanify/chanify/pb"
	"google.golang.org/protobuf/proto"
)

type Message struct {
	pb.Message
}

func NewMessage(tk *Token) *Message {
	m := &Message{}
	m.From = tk.GetNodeID()
	m.Channel = tk.GetChannel()
	return m
}

func (m *Message) DisableToken() *Message {
	m.From = nil
	m.Channel = nil
	return m
}

func (m *Message) TextContent(text string, title string) *Message {
	ctx := &pb.MsgContent{
		Type: pb.MsgType_Text,
		Text: text,
	}
	if len(title) > 0 {
		ctx.Title = title
	}
	m.Content, _ = proto.Marshal(ctx)
	return m
}

func (m *Message) ImageContent(path string) *Message {
	ctx := &pb.MsgContent{
		Type: pb.MsgType_Image,
		File: path,
	}
	m.Content, _ = proto.Marshal(ctx)
	return m
}

func (m *Message) SoundName(sound string) *Message {
	if len(sound) > 0 {
		m.Sound = &pb.Sound{Name: sound}
	}
	return m
}

func (m *Message) SetPriority(priority int) *Message {
	if priority > 0 {
		m.Priority = int32(priority)
	}
	return m
}

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

func (m *Message) Marshal() []byte {
	data, _ := proto.Marshal(&m.Message)
	return data
}
