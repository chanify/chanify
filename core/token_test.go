package core

import (
	"testing"

	"github.com/chanify/chanify/logic"
)

func TestToken(t *testing.T) {
	token := "CJWDwoIGIgIIAQ.dGV4dA"
	tk, err := NewToken(token)
	if err != nil {
		t.Error("Create token failed:", err)
	}
	if tk.String() != token {
		t.Error("Parse token failed")
	}
	if len(tk.sign) <= 0 {
		t.Error("Check token sign failed")
	}
	if _, err := NewToken("CJWDwoIGIgIIAQ"); err != logic.ErrInvalidToken {
		t.Error("Check token format failed")
	}
	if _, err := NewToken("CJWDwoIGIgIIAQ.**"); err == nil {
		t.Error("Check sign format failed")
	}
	if _, err := NewToken("**.--"); err == nil {
		t.Error("Check token data format failed")
	}
	if _, err := NewToken("AJWDwoIGIgIIAQ.--"); err == nil {
		t.Error("Check token proto data failed")
	}
}
