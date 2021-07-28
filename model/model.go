package model

import (
	"errors"
	"strings"

	"github.com/chanify/chanify/pb"
	"google.golang.org/protobuf/proto"
)

// DB base interface
type DB interface {
	GetOption(key string, value interface{}) error
	SetOption(key string, value interface{}) error
	GetUser(uid string) (*User, error)
	UpsertUser(u *User) error
	BindDevice(uid string, uuid string, key []byte, devType int) error
	UnbindDevice(uid string, uuid string) error
	UpdatePushToken(uid string, uuid string, token []byte, sandbox bool) error
	GetDeviceKey(uuid string) ([]byte, error)
	GetDevices(uid string) ([]*Device, error)
	Close()
}

// OpenDB is the function of creating DB instance
type OpenDB func(dsn string) (DB, error)

// variable define
var (
	drivers         = map[string]OpenDB{}
	defaultChannel  []byte
	timelineChannel []byte

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
	timelineChannel, _ = proto.Marshal(&pb.Channel{
		Type: pb.ChanType_Sys,
		Code: pb.ChanCode_TimeSets,
	})
}

// InitDB with DSN
func InitDB(dsn string) (DB, error) {
	dsnItems := strings.Split(dsn, "://")
	if len(dsnItems) <= 1 {
		return nil, ErrInvalidDSN
	}
	dbOpen, ok := drivers[strings.ToLower(dsnItems[0])]
	if !ok {
		return nil, ErrDriverNotFound
	}
	return dbOpen(dsn)
}
