package core

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestUserKey(t *testing.T) {
	c := New("", "", "")
	c.SetSecret("123")
	if _, err := c.GetUserKey("!"); err == nil {
		t.Error("Check user key failed")
	}
	e, _ := hex.DecodeString("a4dc09ee659418e6dad124611bfa3d124cd35124ef401f405f4efc029da9e0fa4bc64725cecaa8ebde9f29ef5db8c7ce45f78e12efa04c4a2e6d8d5a2219ae1e")
	d, err := c.GetUserKey("ABC")
	if err != nil {
		t.Error("Get user key failed:", err)
	}
	if !bytes.Equal(d, e) {
		t.Error("User key is equal")
	}
}
