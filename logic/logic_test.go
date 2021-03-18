package logic

import (
	"testing"
)

func TestLogic(t *testing.T) {
	l, err := NewLogic(&Options{
		DBUrl:  "sqlite://?mode=memory",
		Secret: "123",
	})
	if err != nil {
		t.Fatal("New logic failed:", err)
	}
	defer l.Close()
}

func TestLogicServerless(t *testing.T) {
	l, err := NewLogic(&Options{Secret: "123"})
	if err != nil {
		t.Fatal("New logic failed:", err)
	}
	defer l.Close()
}

func TestLogicFailed(t *testing.T) {
	if _, err := NewLogic(&Options{DBUrl: "nodriver://"}); err == nil {
		t.Fatal("Check logic dburl failed")
	}
	if _, err := NewLogic(&Options{}); err == nil {
		t.Fatal("Check logic secret failed")
	}
}

func TestUser(t *testing.T) {
	l, _ := NewLogic(&Options{Secret: "123"})
	if _, err := l.GetUserKey("GEZDG"); err != nil {
		t.Fatal("Get user key failed")
	}
	if _, err := l.GetUserKey("123"); err == nil {
		t.Fatal("Check get user key failed")
	}
}
