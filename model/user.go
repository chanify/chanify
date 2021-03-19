package model

import "github.com/chanify/chanify/crypto"

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

func (u *User) PublicKeyEncrypt(data []byte) []byte {
	pk, err := crypto.LoadPublicKey(u.PublicKey)
	if err != nil {
		return []byte{}
	}
	out, _ := pk.Encrypt(data) // nolint: errcheck
	return out
}
