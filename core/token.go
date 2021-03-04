package core

import (
	"errors"
	"strings"

	"github.com/chanify/chanify/pb"
	"google.golang.org/protobuf/proto"
)

type Token struct {
	pb.Token
	val string
}

func (c *Core) ParseToken(token string) (*Token, error) {
	tk := &Token{
		val: token,
	}
	tks := strings.Split(token, ".")
	if len(tks) < 2 {
		return nil, errors.New("Invalid token")
	}
	d, err := base64Encode.DecodeString(tks[0])
	if err != nil {
		return nil, err
	}
	if err := proto.Unmarshal(d, &tk.Token); err != nil {
		return nil, err
	}
	return tk, nil
}

func (t *Token) String() string {
	return t.val
}
