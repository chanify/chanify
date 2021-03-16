package core

import (
	"strings"

	"github.com/chanify/chanify/logic"
	"github.com/chanify/chanify/pb"
	"google.golang.org/protobuf/proto"
)

type Token struct {
	pb.Token
	sign []byte
	val  string
}

func NewToken(token string) (*Token, error) {
	tk := &Token{
		val: token,
	}
	tks := strings.Split(token, ".")
	if len(tks) < 2 {
		return nil, logic.ErrInvalidToken
	}
	d, err := base64Encode.DecodeString(tks[0])
	if err != nil {
		return nil, err
	}
	if err := proto.Unmarshal(d, &tk.Token); err != nil {
		return nil, err
	}
	if tk.sign, err = base64Encode.DecodeString(tks[1]); err != nil {
		return nil, err
	}
	return tk, nil
}

func (t *Token) String() string {
	return t.val
}
