package model

import (
	"testing"
)

func TestParseToken(t *testing.T) {
	tk, err := ParseToken("EgMxMjMiBGNoYW4qBU1GUkdH..c2lnbg")
	if err != nil {
		t.Fatal("Parse token failed:", err)
	}
	if tk.GetUserID() != "123" || string(tk.GetNodeID()) != "abc" || string(tk.GetChannel()) != "chan" || tk.RawToken() != "EgMxMjMiBGNoYW4qBU1GUkdH..c2lnbg" {
		t.Fatal("Parse token value failed", string(tk.GetNodeID()))
	}
}

func TestParseTokenFailed(t *testing.T) {
	if _, err := ParseToken("****."); err == nil {
		t.Fatal("Check parse token failed")
	}
	if _, err := ParseToken("***.."); err == nil {
		t.Fatal("Check parse token data failed")
	}
	if _, err := ParseToken("1gMxMjMiBGNoYW4qBU1GUkdH.***."); err == nil {
		t.Fatal("Check parse token decode failed")
	}
	if _, err := ParseToken("EgMxMjMiBGNoYW4qBU1GUkdH.***."); err == nil {
		t.Fatal("Check parse token sign failed")
	}
	if _, err := ParseToken("EgMxMjMiBGNoYW4qBU1GUkdH.c2lnbg.***"); err == nil {
		t.Fatal("Check parse node token format failed")
	}
	tk := &Token{}
	tk.data.NodeId = "***"
	if len(tk.GetNodeID()) > 0 {
		t.Fatal("Check token get node id failed")
	}
	if tk.GetChannel() != nil {
		t.Fatal("Check token get node id failed")
	}
}
