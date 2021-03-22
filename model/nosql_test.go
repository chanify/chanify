package model

import (
	"encoding/hex"
	"testing"
)

func TestNoSQL(t *testing.T) {
	db, err := drivers["nosql"]("nosql://?secret=123456")
	if err != nil {
		t.Fatal("Open nosql failed")
	}
	db.Close()
	var secret []byte
	if err := db.GetOption("secret", &secret); err != nil {
		t.Fatal("GetOption secret failed:", err)
	}
	if err := db.GetOption("name", nil); err != ErrNotImplemented {
		t.Fatal("GetOption failed:", err)
	}
	if err := db.SetOption("name", nil); err != ErrNotImplemented {
		t.Fatal("SetOption failed:", err)
	}
	if err := db.UpsertUser(nil); err != ErrNotImplemented {
		t.Fatal("UpsertUser failed:", err)
	}
	if _, err := db.GetUser("**"); err == nil {
		t.Fatal("Check GetUser failed")
	}
	usr, err := db.GetUser("GEZDG")
	if err != nil {
		t.Fatal("GetUser failed:", err)
	}
	if hex.EncodeToString(usr.SecretKey) != "93c4676a48dbb49dd101d5792b5e023b19abddc6fbbdfc573f8da761b3abd95fdfa1e23102bacfa6090fdc6a4032bb72e28a465890a82939ee088187ce01f594" {
		t.Error("Get user key failed")
	}
	if err := db.BindDevice("", "", nil); err != ErrNotImplemented {
		t.Fatal("Check BindDevice failed:", err)
	}
	if err := db.UnbindDevice("", ""); err != ErrNotImplemented {
		t.Fatal("Check UnbindDevice failed:", err)
	}
	if err := db.UpdatePushToken("", "", nil, false); err != ErrNotImplemented {
		t.Fatal("Check UpdatePushToken failed:", err)
	}
	if _, err := db.GetDeviceKey(""); err != ErrNotImplemented {
		t.Fatal("Check GetDevices failed:", err)
	}
	if _, err := db.GetDevices(""); err != ErrNotImplemented {
		t.Fatal("Check GetDevices failed:", err)
	}

}

func TestNoSQLFailed(t *testing.T) {
	_, err := drivers["nosql"]("nosql://?secret=")
	if err == nil {
		t.Fatal("Check open nosql failed")
	}
}
