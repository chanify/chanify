package model

import (
	"strings"

	"github.com/chanify/chanify/pb"
	"google.golang.org/protobuf/proto"
)

type Token struct {
	data pb.Token
	sign []byte
	raw  string
}

func NewToken() *Token {
	return &Token{}
}

func ParseToken(token string) (*Token, error) {
	tks := strings.Split(token, ".")
	if len(tks) < 2 {
		return nil, ErrInvalidToken
	}
	data, err := base64Encode.DecodeString(tks[0])
	if err != nil {
		return nil, err
	}
	tk := &Token{raw: token}
	if err := proto.Unmarshal(data, &tk.data); err != nil {
		return nil, err
	}
	if tk.sign, err = base64Encode.DecodeString(tks[1]); err != nil {
		return nil, err
	}
	return tk, nil
}

func (tk *Token) GetUserID() string {
	return tk.data.UserId
}

func (tk *Token) GetNodeID() []byte {
	nid, err := base32Encode.DecodeString(tk.data.NodeId)
	if err != nil {
		return []byte{}
	}
	return nid
}

func (tk *Token) GetChannel() []byte {
	if len(tk.data.Channel) <= 0 {
		return defaultChannel
	}
	return tk.data.Channel
}

func (tk *Token) RawToken() string {
	return tk.raw
}
