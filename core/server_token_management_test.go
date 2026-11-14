package core

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chanify/chanify/crypto"
	"github.com/chanify/chanify/logic"
	"github.com/chanify/chanify/model"
	"github.com/chanify/chanify/pb"
	"google.golang.org/protobuf/proto"
)

func TestServerTokenManagementLifecycle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "node.db")
	opts := &logic.Options{DBUrl: "sqlite://" + path, Registerable: true}
	c := New()
	if err := c.Init(opts); err != nil {
		t.Fatal(err)
	}
	defer func() { c.Close() }()
	key := crypto.GenerateSecretKey([]byte("token-management-account"))
	uid := key.ToID(0)
	user, err := c.logic.UpsertUser(uid, key.EncodePublicKey(), false)
	if err != nil {
		t.Fatal(err)
	}
	makeToken := func(id byte) string {
		identifier := bytes.Repeat([]byte{id}, 16)
		body, _ := proto.Marshal(&pb.Token{Scope: pb.TokenScope_ServerSend, TokenId: identifier, UserId: uid, NodeId: c.logic.NodeID, Expires: uint64(time.Now().Add(time.Hour).Unix())})
		mac := hmac.New(sha256.New, user.SecretKey[:32])
		_, _ = mac.Write(body)
		return "cst1:" + crypto.Base64Encode.EncodeToString(body) + ".." + crypto.Base64Encode.EncodeToString(mac.Sum(nil))
	}
	tokens := []string{makeToken(1), makeToken(2)}
	invoke := func(params map[string]interface{}, signingKey *crypto.SecretKey) (int, []*model.ServerTokenRecord) {
		t.Helper()
		params["user"] = uid
		if _, ok := params["nonce"]; !ok {
			params["nonce"] = time.Now().UnixMilli()
		}
		body, _ := json.Marshal(params)
		req := httptest.NewRequest("POST", "/rest/v1/server-tokens", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if signingKey != nil {
			signature, _ := signingKey.Sign(body)
			req.Header.Set("CHUserSign", crypto.Base64Encode.EncodeToString(signature))
		}
		w := httptest.NewRecorder()
		c.APIHandler().ServeHTTP(w, req)
		var result struct {
			Tokens []*model.ServerTokenRecord `json:"tokens"`
		}
		if w.Code == 200 {
			if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
		}
		return w.Code, result.Tokens
	}
	if status, _ := invoke(map[string]interface{}{"action": "list"}, nil); status != 401 {
		t.Fatal("unsigned management accepted")
	}
	other := crypto.GenerateSecretKey([]byte("other-account"))
	if status, _ := invoke(map[string]interface{}{"action": "list"}, other); status != 401 {
		t.Fatal("wrong account signature accepted")
	}
	if status, _ := invoke(map[string]interface{}{"action": "list", "nonce": 1}, key); status != 401 {
		t.Fatal("stale management request accepted")
	}
	if status, records := invoke(map[string]interface{}{"action": "list"}, key); status != 200 || len(records) != 0 {
		t.Fatal("expected no default token")
	}
	for _, raw := range tokens {
		if _, err := c.parseToken(raw); err == nil {
			t.Fatal("unregistered token accepted")
		}
	}
	var records []*model.ServerTokenRecord
	for i, raw := range tokens {
		status, list := invoke(map[string]interface{}{"action": "create", "name": "Integration", "token": raw}, key)
		if status != 200 || len(list) != i+1 {
			t.Fatalf("create %d: %d", i, status)
		}
		records = list
		if _, err := c.parseToken(raw); err != nil {
			t.Fatal("registered token rejected", err)
		}
	}
	tk, _ := model.ParseToken(tokens[0])
	id := tk.ServerTokenID()
	var expiry uint64
	for _, record := range records {
		if record.ID == id {
			expiry = record.Expires
			encrypted, err := crypto.Base64Encode.DecodeString(record.EncryptedToken)
			if err != nil {
				t.Fatal(err)
			}
			plaintext, err := key.Decrypt(encrypted)
			if err != nil || string(plaintext) != tokens[0] {
				t.Fatal("encrypted copy does not round trip", err)
			}
			if strings.Contains(record.EncryptedToken, tokens[0]) {
				t.Fatal("plaintext stored in list")
			}
		}
	}
	if status, _ := invoke(map[string]interface{}{"action": "create", "name": "Retry", "token": tokens[0]}, key); status != 200 {
		t.Fatal("registration retry failed")
	}
	status, records := invoke(map[string]interface{}{"action": "rename", "id": id, "name": "Renamed", "expires": 1, "token": tokens[1]}, key)
	if status != 200 || len(records) != 2 {
		t.Fatal("rename failed")
	}
	for _, record := range records {
		if record.ID == id && (record.Name != "Renamed" || record.Expires != expiry) {
			t.Fatal("rename altered immutable fields")
		}
	}
	if _, err := c.parseToken(tokens[0]); err != nil {
		t.Fatal("rename invalidated bearer")
	}
	c.Close()
	c = New()
	if err := c.Init(opts); err != nil {
		t.Fatal(err)
	}
	if _, err := c.parseToken(tokens[0]); err != nil {
		t.Fatal("registration did not survive restart")
	}
	status, records = invoke(map[string]interface{}{"action": "revoke", "id": id}, key)
	if status != 200 || len(records) != 1 {
		t.Fatal("revoke failed")
	}
	if _, err := c.parseToken(tokens[0]); err == nil {
		t.Fatal("revoked bearer accepted")
	}
	if _, err := c.parseToken(tokens[1]); err != nil {
		t.Fatal("revoke affected another token")
	}
	if status, _ := invoke(map[string]interface{}{"action": "create", "name": "Resurrect", "token": tokens[0]}, key); status != 400 {
		t.Fatal("revoked token was resurrected")
	}
	c.Close()
	c = New()
	if err := c.Init(opts); err != nil {
		t.Fatal(err)
	}
	if _, err := c.parseToken(tokens[0]); err == nil {
		t.Fatal("revocation lost on restart")
	}
	if err := c.logic.DeleteUser(uid); err != nil {
		t.Fatal(err)
	}
	// Rebinding the same identity must not recover records from the deleted account.
	if _, err := c.logic.UpsertUser(uid, key.EncodePublicKey(), false); err != nil {
		t.Fatal(err)
	}
	if status, records := invoke(map[string]interface{}{"action": "list"}, key); status != 200 || len(records) != 0 {
		t.Fatal("deleted account retained token records")
	}

	if _, err := c.parseToken(tokens[1]); err == nil {
		t.Fatal("deleted account token accepted")
	}
}
