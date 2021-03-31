package model

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/subtle"
	"strings"
	"time"

	"github.com/chanify/chanify/pb"
	"google.golang.org/protobuf/proto"
)

type Token struct {
	data     pb.Token
	signSys  []byte
	signNode []byte
	rawData  []byte
	raw      string
}

func ParseToken(token string) (*Token, error) {
	tks := strings.Split(token, ".")
	if len(tks) < 3 {
		return nil, ErrInvalidToken
	}
	data, err := base64Encode.DecodeString(tks[0])
	if err != nil {
		return nil, err
	}
	tk := &Token{raw: token, rawData: data}
	if err := proto.Unmarshal(data, &tk.data); err != nil {
		return nil, err
	}
	if tk.signSys, err = base64Encode.DecodeString(tks[1]); err != nil {
		return nil, err
	}
	if tk.signNode, err = base64Encode.DecodeString(tks[2]); err != nil {
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

func (tk *Token) IsExpires() bool {
	return time.Now().UTC().UnixNano()/1e9 >= int64(tk.data.Expires)
}

func (tk *Token) VerifySign(key []byte) bool {
	mac := hmac.New(sha256.New, key[0:32])
	mac.Write(tk.rawData) // nolint: errcheck
	return hmac.Equal(mac.Sum(nil), tk.signNode)
}

func (tk *Token) VerifyDataHash(data []byte) bool {
	if len(tk.data.DataHash) > 0 && len(data) > 0 {
		h := sha1.Sum(data)
		return subtle.ConstantTimeCompare(tk.data.DataHash, h[:]) == 1
	}
	return false
}

func (tk *Token) RawToken() string {
	return tk.raw
}
