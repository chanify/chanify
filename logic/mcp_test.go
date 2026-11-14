package logic

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/chanify/chanify/crypto"
	"github.com/chanify/chanify/model"
)

func TestMCPCredentialLifecycle(t *testing.T) {
	opts := &Options{DBUrl: "sqlite://" + filepath.Join(t.TempDir(), "node.db"), Registerable: true}
	l, err := NewLogic(opts)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { l.Close() }()
	key := crypto.GenerateSecretKey([]byte("synthetic-mcp-identity"))
	uid := key.ToID(0)
	if _, err := l.UpsertUser(uid, key.EncodePublicKey(), false); err != nil {
		t.Fatal(err)
	}
	r, raw, err := l.CreateMCPToken(uid, "Agent", time.Now().Add(time.Hour).Unix())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := model.ParseToken(raw); err == nil {
		t.Fatal("MCP bearer parsed as sending token")
	}
	second, raw2, err := l.CreateMCPToken(uid, "Second", time.Now().Add(time.Hour).Unix())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := l.AuthenticateMCP(raw); err != nil {
		t.Fatal(err)
	}
	if err := l.UpdateMCPToken(uid, r.ID, "Renamed", false); err != nil {
		t.Fatal(err)
	}
	if _, err := l.AuthenticateMCP(raw); err != nil {
		t.Fatal("rename changed credential")
	}
	if err := l.UpdateMCPToken("another-user", r.ID, "", true); err == nil {
		t.Fatal("cross user revocation accepted")
	}
	l.Close()
	l, err = NewLogic(opts)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := l.AuthenticateMCP(raw); err != nil {
		t.Fatal("restart lost credential", err)
	}
	if err := l.UpdateMCPToken(uid, r.ID, "", true); err != nil {
		t.Fatal(err)
	}
	if _, err := l.AuthenticateMCP(raw); err == nil {
		t.Fatal("revoked credential accepted")
	}
	if _, err := l.AuthenticateMCP(raw2); err != nil {
		t.Fatal("other credential affected")
	}
	if err := l.UpdateMCPToken(uid, second.ID, "", true); err != nil {
		t.Fatal(err)
	}
	if _, _, err := l.CreateMCPToken(uid, "Invalid", 1); err == nil {
		t.Fatal("expired credential created")
	}
	if _, _, err := l.CreateMCPToken(uid, "Invalid", time.Now().Add(time.Hour).Unix()); err != nil {
		t.Fatal("credential must not require channels", err)
	}
	if err := l.DeleteUser(uid); err != nil {
		t.Fatal(err)
	}
	list, err := l.ListMCPTokens(uid)
	if err != nil || len(list) != 0 {
		t.Fatal("account deletion retained credentials", err)
	}
}
