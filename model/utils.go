package model

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/chanify/chanify/crypto"
)

func NewAESGCM(key []byte) (cipher.AEAD, error) {
	if len(key) < 32 {
		return nil, errors.New("invalid key")
	}
	block, _ := aes.NewCipher(key[:32])
	return cipher.NewGCM(block)
}

func DecodePushToken(token string) ([]byte, error) {
	return Base64Encode.DecodeString(token)
}

func CalcDeviceKey(uuid string, key string) (*crypto.PublicKey, error) {
	data, err := Base64Encode.DecodeString(key)
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
	data, err := Base64Encode.DecodeString(key)
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
	if Base32Encode.EncodeToString(udata) != uid {
		return nil, ErrInvalidUserID
	}
	return pk, nil
}
