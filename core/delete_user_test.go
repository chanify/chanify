package core

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/chanify/chanify/crypto"
	"github.com/chanify/chanify/logic"
)

func TestDeleteUserOwnershipAndRetry(t *testing.T) {
	c := New()
	if err := c.Init(&logic.Options{DBUrl: "sqlite://?mode=memory", Registerable: true}); err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	owner := crypto.GenerateSecretKey(nil)
	stranger := crypto.GenerateSecretKey(nil)
	uid := owner.ToID(0)
	if _, err := c.logic.UpsertUser(uid, owner.EncodePublicKey(), false); err != nil {
		t.Fatal(err)
	}
	handler := c.APIHandler()
	for _, tc := range []struct {
		name   string
		key    *crypto.SecretKey
		nonce  int64
		status int
	}{
		{"unsigned owner substitution", stranger, time.Now().UnixMilli(), http.StatusUnauthorized},
		{"expired signature", owner, time.Now().Add(-10 * time.Minute).UnixMilli(), http.StatusUnauthorized},
		{"owner", owner, time.Now().UnixMilli(), http.StatusOK},
		{"retry", owner, time.Now().UnixMilli(), http.StatusOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body, err := json.Marshal(map[string]interface{}{"user": uid, "key": owner.EncodePublicKey(), "nonce": tc.nonce})
			if err != nil {
				t.Fatal(err)
			}
			sig, err := tc.key.Sign(body)
			if err != nil {
				t.Fatal(err)
			}
			req := httptest.NewRequest("POST", "/rest/v1/delete-user", strings.NewReader(string(body)))
			req.Header.Set("CHUserSign", crypto.Base64Encode.EncodeToString(sig))
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			if w.Code != tc.status {
				t.Fatalf("status %d, want %d", w.Code, tc.status)
			}
			if tc.status != http.StatusOK {
				if _, err := c.logic.GetUser(uid); err != nil {
					t.Fatal("unauthorized request removed account")
				}
			}
		})
	}
}
