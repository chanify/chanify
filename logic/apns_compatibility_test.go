package logic

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"strings"
	"testing"

	"github.com/chanify/chanify/crypto"
	"github.com/sideshow/apns2/token"
)

// Exercise APNS's JWT signer with an existing P-256 key after the SDK/JWT upgrade.
func TestAPNSAuthenticationKeyCompatibility(t *testing.T) {
	key := crypto.GenerateSecretKey([]byte("123"))
	auth := &token.Token{AuthKey: &key.PrivateKey, KeyID: "test-key", TeamID: "test-team"}
	ok, err := auth.Generate()
	if err != nil || !ok {
		t.Fatalf("APNS authentication token failed: %v", err)
	}
	parts := strings.Split(auth.Bearer, ".")
	if len(parts) != 3 {
		t.Fatal("invalid JWT format")
	}
	decode := func(s string) []byte {
		t.Helper()
		b, err := base64.RawURLEncoding.DecodeString(s)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	var header struct {
		Algorithm string `json:"alg"`
		KeyID     string `json:"kid"`
	}
	if err := json.Unmarshal(decode(parts[0]), &header); err != nil {
		t.Fatal(err)
	}
	if header.Algorithm != "ES256" || header.KeyID != "test-key" {
		t.Fatal("APNS JWT header changed")
	}
	var claims struct {
		Issuer   string `json:"iss"`
		IssuedAt int64  `json:"iat"`
	}
	if err := json.Unmarshal(decode(parts[1]), &claims); err != nil {
		t.Fatal(err)
	}
	if claims.Issuer != "test-team" || claims.IssuedAt <= 0 {
		t.Fatal("APNS JWT claims changed")
	}
	sig := decode(parts[2])
	if len(sig) != 64 {
		t.Fatal("APNS ES256 signature format changed")
	}
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if !ecdsa.Verify(&key.PublicKey, digest[:], new(big.Int).SetBytes(sig[:32]), new(big.Int).SetBytes(sig[32:])) {
		t.Fatal("APNS JWT signature rejected")
	}
}
