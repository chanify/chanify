package model

import (
	"encoding/base32"
	"errors"
	"net/url"
	"strings"
)

type DB interface {
	GetOption(key string, value interface{}) error
	SetOption(key string, value interface{}) error
	GetUser(uid string) (*User, error)
	UpsertUser(u *User) error
	Close()
}

type OpenDB func(dsn *url.URL) (DB, error)

var (
	drivers      = map[string]OpenDB{}
	base32Encode = base32.StdEncoding.WithPadding(base32.NoPadding)

	ErrDriverNotFound = errors.New("driver not found")
	ErrNotImplemented = errors.New("not implemented")
	ErrInvalidDSN     = errors.New("invalid dsn")
)

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
