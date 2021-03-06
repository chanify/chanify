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

// NewAESGCM for aes-gcm chiper
func NewAESGCM(key []byte) (cipher.AEAD, error) {
	if len(key) < 32 {
		return nil, errors.New("invalid key")
	}
	block, _ := aes.NewCipher(key[:32])
	return cipher.NewGCM(block)
}

// DecodePushToken for APNS
func DecodePushToken(token string) ([]byte, error) {
	return crypto.Base64Encode.DecodeString(token)
}

// CalcDeviceKey from device public key
func CalcDeviceKey(uuid string, key string) (*crypto.PublicKey, error) {
	data, err := crypto.Base64Encode.DecodeString(key)
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

// CalcUserKey from user public key
func CalcUserKey(uid string, key string) (*crypto.PublicKey, error) {
	data, err := crypto.Base64Encode.DecodeString(key)
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
	if crypto.Base32Encode.EncodeToString(udata) != uid {
		return nil, ErrInvalidUserID
	}
	return pk, nil
}
