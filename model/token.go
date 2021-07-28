package model

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/subtle"
	"strings"
	"time"

	"github.com/chanify/chanify/crypto"
	"github.com/chanify/chanify/pb"
	"google.golang.org/protobuf/proto"
)

// Token for sender
type Token struct {
	data     pb.Token
	signSys  []byte
	signNode []byte
	rawData  []byte
	raw      string
}

// ParseToken create token from base64 string
func ParseToken(token string) (*Token, error) {
	tks := strings.Split(token, ".")
	if len(tks) < 3 {
		return nil, ErrInvalidToken
	}
	data, err := crypto.Base64Encode.DecodeString(tks[0])
	if err != nil {
		return nil, err
	}
	tk := &Token{raw: token, rawData: data}
	if err := proto.Unmarshal(data, &tk.data); err != nil {
		return nil, err
	}
	if tk.signSys, err = crypto.Base64Encode.DecodeString(tks[1]); err != nil {
		return nil, err
	}
	if tk.signNode, err = crypto.Base64Encode.DecodeString(tks[2]); err != nil {
		return nil, err
	}
	return tk, nil
}

// GetUserID return user id string
func (tk *Token) GetUserID() string {
	return tk.data.UserId
}

// GetNodeID return node id
func (tk *Token) GetNodeID() []byte {
	nid, err := crypto.Base32Encode.DecodeString(tk.data.NodeId)
	if err != nil {
		return []byte{}
	}
	return nid
}

// GetChannel return channel code
func (tk *Token) GetChannel() []byte {
	return tk.data.Channel
}

// IsExpires check token expires timestamp(UTC)
func (tk *Token) IsExpires() bool {
	return time.Now().UTC().UnixNano()/1e9 >= int64(tk.data.Expires)
}

// VerifySign check token sign
func (tk *Token) VerifySign(key []byte) bool {
	mac := hmac.New(sha256.New, key[0:32])
	mac.Write(tk.rawData) // nolint: errcheck
	return hmac.Equal(mac.Sum(nil), tk.signNode)
}

// VerifyDataHash check the hash of uri limit
func (tk *Token) VerifyDataHash(data []byte) bool {
	if len(tk.data.DataHash) > 0 && len(data) > 0 {
		h := sha1.Sum(data)
		return subtle.ConstantTimeCompare(tk.data.DataHash, h[:]) == 1
	}
	return false
}

// RawToken return raw value
func (tk *Token) RawToken() string {
	return tk.raw
}

// HashValue return sha1 with token raw value
func (tk *Token) HashValue() []byte {
	h := sha1.Sum([]byte(tk.raw))
	return h[:]
}
