package model

import "github.com/chanify/chanify/crypto"

// User infomation
type User struct {
	UID       string
	PublicKey []byte
	SecretKey []byte
	Flags     uint
}

// IsServerless for user configuration
func (u *User) IsServerless() bool {
	return (u.Flags&0x01 == 0)
}

// SetServerless for user configuration
func (u *User) SetServerless(s bool) {
	if s {
		u.Flags &= ^uint(0x01)
	} else {
		u.Flags |= uint(0x01)
	}
}

// GetPublicKeyString return the user public key
func (u *User) GetPublicKeyString() string {
	return crypto.Base64Encode.EncodeToString(u.PublicKey)
}

// PublicKeyEncrypt return encrypted public key
func (u *User) PublicKeyEncrypt(data []byte) []byte {
	pk, err := crypto.LoadPublicKey(u.PublicKey)
	if err != nil {
		return []byte{}
	}
	out, _ := pk.Encrypt(data) // nolint: errcheck
	return out
}
