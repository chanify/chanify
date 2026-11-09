package logic

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/chanify/chanify/crypto"
	"github.com/chanify/chanify/model"
)

func TestStoredNodeKeySurvivesRestart(t *testing.T) {
	dsn := "sqlite://" + filepath.Join(t.TempDir(), "node.db")
	key := crypto.GenerateSecretKey([]byte("123"))
	pem := key.MarshalSecretKey()
	db, err := model.InitDB(dsn)
	if err != nil {
		t.Fatal(err)
	}
	err = db.SetOption("secret", pem)
	db.Close()
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		l, err := NewLogic(&Options{DBUrl: dsn})
		if err != nil {
			t.Fatal(err)
		}
		if l.NodeID != "AGIEVPVLCVE4HSEYWQC6GCSZYUWXNPTCDA" {
			t.Error("stored legacy node ID changed")
		}
		var saved []byte
		err = l.db.GetOption("secret", &saved)
		l.Close()
		if err != nil || !bytes.Equal(saved, pem) {
			t.Fatalf("stored key changed: %v", err)
		}
	}
}

func TestInvalidStoredNodeKeyIsNotReplaced(t *testing.T) {
	dsn := "sqlite://" + filepath.Join(t.TempDir(), "node.db")
	invalid := []byte("invalid stored private key")
	db, err := model.InitDB(dsn)
	if err != nil {
		t.Fatal(err)
	}
	err = db.SetOption("secret", invalid)
	db.Close()
	if err != nil {
		t.Fatal(err)
	}
	if l, err := NewLogic(&Options{DBUrl: dsn}); err == nil {
		l.Close()
		t.Fatal("invalid stored key silently replaced")
	}
	db, err = model.InitDB(dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var saved []byte
	if err := db.GetOption("secret", &saved); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(saved, invalid) {
		t.Fatal("invalid stored key overwritten")
	}
}
