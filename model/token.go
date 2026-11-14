package model

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

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
	channel  []byte
}

// ParseToken create token from base64 string
func ParseToken(token string) (*Token, error) {
	server := strings.HasPrefix(token, "cst1:")
	raw := token
	token = strings.TrimPrefix(token, "cst1:")
	tks := strings.Split(token, ".")
	if len(tks) < 3 {
		return nil, ErrInvalidToken
	}
	data, err := crypto.Base64Encode.DecodeString(tks[0])
	if err != nil {
		return nil, err
	}
	tk := &Token{raw: raw, rawData: data}
	if err := proto.Unmarshal(data, &tk.data); err != nil {
		return nil, err
	}
	if tk.signSys, err = crypto.Base64Encode.DecodeString(tks[1]); err != nil {
		return nil, err
	}
	if tk.signNode, err = crypto.Base64Encode.DecodeString(tks[2]); err != nil {
		return nil, err
	}
	if server {
		if len(tks) != 3 || tk.data.Scope != pb.TokenScope_ServerSend || len(tk.data.TokenId) != 16 || tk.data.NodeId == "" || tk.data.UserId == "" || len(tk.data.Channel) != 0 || len(tk.data.DeviceId) != 0 || len(tk.data.DataHash) != 0 || len(tk.signSys) != 0 || len(tk.signNode) != sha256.Size {
			return nil, ErrInvalidToken
		}
	} else if tk.data.Scope != pb.TokenScope_ChannelSend || len(tk.data.TokenId) != 0 {
		return nil, ErrInvalidToken
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
	if tk.IsServerToken() {
		return tk.channel
	}
	return tk.data.Channel
}

// IsExpires check token expires timestamp(UTC)
func (tk *Token) IsExpires() bool {
	return uint64(time.Now().Unix()) >= tk.data.Expires
}

// VerifySign check token sign
func (tk *Token) VerifySign(key []byte) bool {
	if len(key) < 32 {
		return false
	}
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

// IsServerToken identifies an explicitly scoped server-wide sending credential.
func (tk *Token) IsServerToken() bool { return tk.data.Scope == pb.TokenScope_ServerSend }

// ForChannel creates a per-request destination without changing signed token data.
// Channel is the unpadded base64url-encoded protobuf Channel used by Apple clients.
func (tk *Token) ForChannel(value string) (*Token, error) {
	if !tk.IsServerToken() {
		return tk, nil
	}
	if value == "" {
		return nil, fmt.Errorf("channel is required")
	}
	if len(value) > 1024 {
		return nil, fmt.Errorf("invalid channel")
	}
	data, err := crypto.Base64Encode.DecodeString(value)
	if err != nil || len(data) == 0 {
		return nil, fmt.Errorf("invalid channel")
	}
	var channel pb.Channel
	if proto.Unmarshal(data, &channel) != nil || len(channel.ProtoReflect().GetUnknown()) != 0 {
		return nil, fmt.Errorf("invalid channel")
	}
	switch channel.Type {
	case pb.ChanType_Sys:
		if (channel.Code != pb.ChanCode_Uncategorized && channel.Code != pb.ChanCode_TimeSets) || channel.Name != "" || channel.Icon != "" {
			return nil, fmt.Errorf("invalid channel")
		}
	case pb.ChanType_User:
		if channel.Code != pb.ChanCode_Uncategorized || strings.TrimSpace(channel.Name) == "" || !utf8.ValidString(channel.Name) || len(channel.Name) > 256 || len(channel.Icon) > 256 {
			return nil, fmt.Errorf("invalid channel")
		}
	default:
		return nil, fmt.Errorf("invalid channel")
	}
	// Copy only immutable fields: never copy the protobuf message's internal lock.
	return &Token{raw: tk.raw, rawData: tk.rawData, signSys: tk.signSys, signNode: tk.signNode, channel: data, data: pb.Token{UserId: tk.data.UserId, NodeId: tk.data.NodeId, Scope: tk.data.Scope, Expires: tk.data.Expires, TokenId: tk.data.TokenId}}, nil
}

// ServerTokenID is the stable identifier for registration and revocation.
func (tk *Token) ServerTokenID() string { return hex.EncodeToString(tk.data.TokenId) }
func (tk *Token) ExpiresAt() uint64     { return tk.data.Expires }

// ForOptionalChannel defaults only an omitted destination, never an explicit empty value.
func (tk *Token) ForOptionalChannel(value string, provided bool) (*Token, error) {
	if !provided {
		value = crypto.Base64Encode.EncodeToString(defaultChannel)
	}
	return tk.ForChannel(value)
}
