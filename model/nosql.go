package model

import (
	"crypto/sha256"
	"crypto/sha512"
	"net/url"

	"github.com/chanify/chanify/crypto"
)

type nosql struct {
	secret []byte
	seckey []byte
}

func init() {
	drivers["nosql"] = func(dsn *url.URL) (DB, error) {
		secret := []byte(dsn.Query().Get("secret"))
		if len(secret) <= 0 {
			return nil, ErrInvalidDSN
		}
		return &nosql{
			secret: sha256.New().Sum(secret),
			seckey: crypto.GenerateSecretKey(secret).MarshalSecretKey(),
		}, nil
	}
}

func (s *nosql) Close() {
}

func (s *nosql) GetOption(key string, value interface{}) error {
	if key == "secret" {
		*(value.(*[]byte)) = s.seckey
		return nil
	}
	return ErrNotImplemented
}

func (s *nosql) SetOption(key string, value interface{}) error {
	return ErrNotImplemented
}

func (s *nosql) GetUser(uid string) (*User, error) {
	data, err := base32Encode.DecodeString(uid)
	if err != nil {
		return nil, err
	}
	h := sha512.New()
	h.Write(s.secret) // nolint: errcheck
	h.Write(data)     // nolint: errcheck
	return &User{
		Uid:       uid,
		SecretKey: h.Sum(nil),
	}, nil
}

func (s *nosql) UpsertUser(u *User) error {
	return ErrNotImplemented
}

func (s *nosql) BindDevice(uid string, uuid string, key []byte) error {
	return ErrNotImplemented
}

func (s *nosql) UnbindDevice(uid string, uuid string) error {
	return ErrNotImplemented
}

func (s *nosql) UpdatePushToken(uid string, uuid string, token []byte, sandbox bool) error {
	return ErrNotImplemented
}

func (s *nosql) GetDeviceKey(uuid string) ([]byte, error) {
	return nil, ErrNotImplemented
}

func (s *nosql) GetDevices(uid string) ([]*Device, error) {
	return nil, ErrNotImplemented
}
