package model

import (
	"bytes"
	"testing"
)

func TestNewToken(t *testing.T) {
	tk := NewToken()
	if tk == nil {
		t.Fatal("Create token failed")
	}
}

func TestParseToken(t *testing.T) {
	tk, err := ParseToken("EgMxMjMiBGNoYW4qBU1GUkdH.c2lnbg")
	if err != nil {
		t.Fatal("Parse token failed:", err)
	}
	if tk.GetUserID() != "123" || string(tk.GetNodeID()) != "abc" || string(tk.GetChannel()) != "chan" || tk.RawToken() != "EgMxMjMiBGNoYW4qBU1GUkdH.c2lnbg" {
		t.Fatal("Parse token value failed", string(tk.GetNodeID()))
	}
}

func TestParseTokenFailed(t *testing.T) {
	if _, err := ParseToken("****."); err == nil {
		t.Fatal("Check parse token failed")
	}
	if _, err := ParseToken("EgMxMjMiBGNoYW4qBU1GUkdH.***"); err == nil {
		t.Fatal("Check parse token sign failed")
	}
	if _, err := ParseToken("c2lnbg.***"); err == nil {
		t.Fatal("Check parse token format failed")
	}
	tk := &Token{}
	tk.data.NodeId = "***"
	if len(tk.GetNodeID()) > 0 {
		t.Fatal("Check token get node id failed")
	}
	if !bytes.Equal(tk.GetChannel(), defaultChannel) {
		t.Fatal("Check token get node id failed")
	}
}
