package model

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/chanify/chanify/crypto"
)

type User struct {
	Uid       string
	PublicKey []byte
	SecretKey []byte
	Flags     uint
}

func (u *User) IsServerless() bool {
	return (u.Flags&0x01 == 0)
}

func (u *User) SetServerless(s bool) {
	if s {
		u.Flags &= ^uint(0x01)
	} else {
		u.Flags |= uint(0x01)
	}
}

func DecodePushToken(token string) ([]byte, error) {
	return base64Encode.DecodeString(token)
}

func CalcDeviceKey(uuid string, key string) (*crypto.PublicKey, error) {
	data, err := base64Encode.DecodeString(key)
	if err != nil {
		return nil, err
	}
	pk, err := crypto.LoadPublicKey(data)
	if err != nil {
		return nil, err
	}
	h := sha1.Sum(data)
	if strings.ToUpper(hex.EncodeToString(h[:])) != uuid {
		return nil, ErrInvalidDeviceID
	}
	return pk, nil
}

func CalcUserKey(uid string, key string) (*crypto.PublicKey, error) {
	data, err := base64Encode.DecodeString(key)
	if err != nil {
		return nil, err
	}
	pk, err := crypto.LoadPublicKey(data)
	if err != nil {
		return nil, err
	}
	// Calc user id
	h1 := sha256.Sum256(data)
	out := append([]byte{}, h1[:]...)
	out = append(out, data...)
	h := sha1.Sum(out)
	udata := append([]byte{0x00}, h[:]...)
	if base32Encode.EncodeToString(udata) != uid {
		return nil, ErrInvalidUserID
	}
	return pk, nil
}
