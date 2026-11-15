package logic

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"github.com/chanify/chanify/pb"
	"google.golang.org/protobuf/proto"
	"strings"
	"testing"
	"time"

	"github.com/chanify/chanify/crypto"
	"github.com/chanify/chanify/model"
)

// Embed the real interfaces so each test overrides only the operation under test.
type credentialDB struct {
	model.DB
	model.AdminStore
	model.MCPTokenStore
	model.ServerTokenStore
	user                 *model.User
	userErr, errorResult error
}

func (d *credentialDB) GetUser(string) (*model.User, error)                  { return d.user, d.userErr }
func (d *credentialDB) AdminUsers() ([]model.AdminUser, error)               { return nil, d.errorResult }
func (d *credentialDB) AdminTokens() ([]model.AdminToken, error)             { return nil, d.errorResult }
func (d *credentialDB) CreateMCPToken(*model.MCPTokenRecord) error           { return d.errorResult }
func (d *credentialDB) GetMCPToken(string) (*model.MCPTokenRecord, error)    { return nil, d.errorResult }
func (d *credentialDB) UpdateMCPToken(string, string, string, bool) error    { return d.errorResult }
func (d *credentialDB) UpdateServerToken(string, string, string, bool) error { return d.errorResult }

func TestCredentialStoreUnavailable(t *testing.T) {
	l, err := NewLogic(&Options{Secret: "synthetic"})
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	if l.AdminAvailable() {
		t.Fatal("stateless admin enabled")
	}
	if _, _, err = l.AdminRecords(); !errors.Is(err, model.ErrNotImplemented) {
		t.Fatal(err)
	}
	if err = l.AdminUpdateToken("sending", "u", "id", "name", false); !errors.Is(err, model.ErrNotImplemented) {
		t.Fatal(err)
	}
	if _, _, err = l.CreateMCPToken("u", "name", time.Now().Add(time.Hour).Unix()); !errors.Is(err, model.ErrNotImplemented) {
		t.Fatal(err)
	}
	if _, err = l.ListMCPTokens("u"); !errors.Is(err, model.ErrNotImplemented) {
		t.Fatal(err)
	}
	if err = l.UpdateMCPToken("u", "id", "name", false); !errors.Is(err, model.ErrNotImplemented) {
		t.Fatal(err)
	}
	if _, err = l.ManageServerTokens("u", "list", "", "", ""); !errors.Is(err, model.ErrNotImplemented) {
		t.Fatal(err)
	}
	raw := "cmcp1." + strings.Repeat("a", 32) + "." + strings.Repeat("b", 43)
	if _, err = l.AuthenticateMCP(raw); !errors.Is(err, model.ErrInvalidToken) {
		t.Fatal(err)
	}
}

func TestCredentialValidationAndStoreFailures(t *testing.T) {
	want := errors.New("store failure")
	db := &credentialDB{user: &model.User{UID: "u", Flags: 1}, errorResult: want}
	l := &Logic{db: db}
	if _, _, err := l.AdminRecords(); !errors.Is(err, want) {
		t.Fatal(err)
	}
	if err := l.AdminUpdateToken("unknown", "u", "id", "name", false); err == nil {
		t.Fatal("unknown kind accepted")
	}
	for _, name := range []string{" ", strings.Repeat("a", 129), string([]byte{255})} {
		if err := l.AdminUpdateToken("sending", "u", "id", name, false); err == nil || errors.Is(err, want) {
			t.Fatal(err)
		}
	}
	for _, kind := range []string{"sending", "mcp"} {
		if err := l.AdminUpdateToken(kind, "u", "id", "name", false); !errors.Is(err, want) {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"", strings.Repeat("a", 129)} {
		if err := l.UpdateMCPToken("u", "id", name, false); err == nil || errors.Is(err, want) {
			t.Fatal(err)
		}
	}
	if _, _, err := l.CreateMCPToken("u", "name", time.Now().Add(time.Hour).Unix()); !errors.Is(err, want) {
		t.Fatal(err)
	}
	db.userErr = want
	if _, _, err := l.CreateMCPToken("u", "name", time.Now().Add(time.Hour).Unix()); !errors.Is(err, model.ErrInvalidUserID) {
		t.Fatal(err)
	}
	if _, err := l.ManageServerTokens("u", "list", "", "", ""); !errors.Is(err, model.ErrInvalidUserID) {
		t.Fatal(err)
	}
	db.userErr = nil
	for _, tc := range []struct {
		action, id, name, raw string
		want                  error
	}{
		{"create", "", " ", "", nil}, {"create", "", "name", "invalid", model.ErrInvalidToken},
		{"rename", "bad", "name", "", model.ErrInvalidToken}, {"rename", strings.Repeat("a", 32), "name", "", want},
		{"unknown", "", "", "", model.ErrNotImplemented},
	} {
		_, err := l.ManageServerTokens("u", tc.action, tc.id, tc.name, tc.raw)
		if err == nil || (tc.want != nil && !errors.Is(err, tc.want)) {
			t.Fatal(tc, err)
		}
	}
}

func TestAuthenticateMCPRejectsDeletedAccount(t *testing.T) {
	l, err := NewLogic(&Options{DBUrl: "sqlite://?mode=memory", Registerable: true})
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	key := crypto.GenerateSecretKey([]byte("synthetic-auth-test"))
	uid := key.ToID(0)
	if _, err := l.UpsertUser(uid, key.EncodePublicKey(), false); err != nil {
		t.Fatal(err)
	}
	_, raw, err := l.CreateMCPToken(uid, "Agent", time.Now().Add(time.Hour).Unix())
	if err != nil {
		t.Fatal(err)
	}
	original := l.db
	// Delegate the token lookup while forcing only the user lookup to fail.
	l.db = &missingAccountDB{DB: original, MCPTokenStore: original.(model.MCPTokenStore)}
	if _, err := l.AuthenticateMCP(raw); !errors.Is(err, model.ErrInvalidToken) {
		t.Fatal(err)
	}
	l.db = original
}

type missingAccountDB struct {
	model.DB
	model.MCPTokenStore
}

func (*missingAccountDB) GetUser(string) (*model.User, error) { return nil, model.ErrInvalidUserID }

type userOnlyDB struct {
	model.DB
	user *model.User
}

func (d *userOnlyDB) GetUser(string) (*model.User, error) { return d.user, nil }
func serverCandidate(t *testing.T, node string, user *model.User) *model.Token {
	t.Helper()
	data, err := proto.Marshal(&pb.Token{UserId: user.UID, NodeId: node, Scope: pb.TokenScope_ServerSend, TokenId: make([]byte, 16), Expires: uint64(time.Now().Add(time.Hour).Unix())})
	if err != nil {
		t.Fatal(err)
	}
	mac := hmac.New(sha256.New, user.SecretKey[:32])
	_, _ = mac.Write(data)
	raw := "cst1:" + crypto.Base64Encode.EncodeToString(data) + ".." + crypto.Base64Encode.EncodeToString(mac.Sum(nil))
	tk, err := model.ParseToken(raw)
	if err != nil {
		t.Fatal(err)
	}
	return tk
}
func TestServerTokenStoreAndEncryptionRequired(t *testing.T) {
	user := &model.User{UID: "synthetic", Flags: 1, SecretKey: make([]byte, 64)}
	node := crypto.Base32Encode.EncodeToString([]byte("node"))
	l := &Logic{NodeID: node, db: &userOnlyDB{user: user}}
	tk := serverCandidate(t, node, user)
	if l.VerifyToken(tk) {
		t.Fatal("registered store requirement bypassed")
	}
	l.db = &credentialDB{user: user}
	if _, err := l.ManageServerTokens(user.UID, "create", "", "name", tk.RawToken()); !errors.Is(err, model.ErrInvalidToken) {
		t.Fatal("invalid encryption key accepted", err)
	}
}
