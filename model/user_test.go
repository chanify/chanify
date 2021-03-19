package model

import "testing"

func TestUser(t *testing.T) {
	u := &User{}
	u.SetServerless(false)
	if u.IsServerless() {
		t.Fatal("Set serverless false failed")
	}
	u.SetServerless(true)
	if !u.IsServerless() {
		t.Fatal("Set serverless true failed")
	}
}

func TestPublicKeyEncrypt(t *testing.T) {
	u := &User{}
	if len(u.PublicKeyEncrypt([]byte{})) > 0 {
		t.Fatal("Check public key encrypt failed")
	}
}
