package model

import (
	"testing"
)

func TestToken(t *testing.T) {
	tk := NewToken()
	if tk == nil {
		t.Fatal("Create token failed")
	}
	// tk.expires = time.Now().Add(time.Minute)
	// if !tk.IsExpires() {
	// 	t.Fatal("Check token expires failed")
	// }
	// tk.expires = time.Now().Add(-time.Minute)
	// if tk.IsExpires() {
	// 	t.Fatal("Check invalid token expires failed")
	// }
}
