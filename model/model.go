package model

import (
	"encoding/base32"
	"encoding/base64"
	"errors"
	"net/url"
	"strings"

	"github.com/chanify/chanify/pb"
	"google.golang.org/protobuf/proto"
)

type DB interface {
	GetOption(key string, value interface{}) error
	SetOption(key string, value interface{}) error
	GetUser(uid string) (*User, error)
	UpsertUser(u *User) error
	BindDevice(uid string, uuid string, key []byte) error
	UnbindDevice(uid string, uuid string) error
	UpdatePushToken(uid string, uuid string, token []byte, sandbox bool) error
	GetDevices(uid string) ([]*Device, error)
	Close()
}

type OpenDB func(dsn *url.URL) (DB, error)

var (
	drivers        = map[string]OpenDB{}
	base64Encode   = base64.RawURLEncoding
	base32Encode   = base32.StdEncoding.WithPadding(base32.NoPadding)
	defaultChannel []byte

	ErrDriverNotFound   = errors.New("driver not found")
	ErrNotImplemented   = errors.New("not implemented")
	ErrInvalidDeviceID  = errors.New("invalid device id")
	ErrInvalidUserID    = errors.New("invalid user id")
	ErrInvalidPublicKey = errors.New("invalid public key")
	ErrInvalidToken     = errors.New("invalid token")
	ErrInvalidDSN       = errors.New("invalid dsn")
)

func init() {
	defaultChannel, _ = proto.Marshal(&pb.Channel{
		Type: pb.ChanType_Sys,
		Code: pb.ChanCode_Uncategorized,
	})
}

func InitDB(dsn string) (DB, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}
	dbOpen, ok := drivers[strings.ToLower(u.Scheme)]
	if !ok {
		return nil, ErrDriverNotFound
	}
	return dbOpen(u)
}
