package logic

import (
	"time"
)

type Token struct {
	expires time.Time
}

func NewToken() *Token {
	return &Token{}
}

func (tk *Token) IsExpires() bool {
	return time.Now().Before(tk.expires)
}
