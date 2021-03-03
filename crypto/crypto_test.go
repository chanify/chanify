package crypto

import (
	"bytes"
	"encoding/base64"
	"math/big"
	"testing"
)

func TestGenKey(t *testing.T) {
	k1, err := GenerateSecretKey([]byte("123"))
	if err != nil {
		t.Error("Create secret failed")
	}
	kid := k1.ToID(0x01)
	if len(kid) <= 0 {
		t.Error("To id failed")
	}
	k2, _ := GenerateSecretKey([]byte("123"))
	if kid != k2.ToID(0x01) {
		t.Error("Fix secret key failed")
	}
}

func TestPublicKey(t *testing.T) {
	k, _ := GenerateSecretKey([]byte("123"))
	data := k.EncodePublicKey()
	if len(data) <= 0 {
		t.Error("Encode public key failed")
	}
	dk, err := base64.RawStdEncoding.DecodeString(data)
	if err != nil {
		t.Error("Encode public key invalid:", err)
	}
	k2, err := LoadPublicKey(dk)
	if err != nil {
		t.Error("Load public key failed:", err)
	}
	if k.ToID(0x01) != k2.ToID(0x01) {
		t.Error("Check public key failed")
	}
}

func TestCrypto(t *testing.T) {
	data := []byte("hello")
	k1, _ := GenerateSecretKey([]byte("123"))
	k2 := k1.GetPublicKey()
	d, err := k2.Encrypt(data)
	if err != nil {
		t.Error("Encrypt failed:", err)
	}
	out, err := k1.Decrypt(d)
	if err != nil {
		t.Error("Decrypt failed:", err)
	}
	if !bytes.Equal(data, out) {
		t.Error("Encrypt & decrypt failed")
	}
}

func TestSign(t *testing.T) {
	data := []byte("hello")
	k1, _ := GenerateSecretKey([]byte("123"))
	k2 := k1.GetPublicKey()
	s, err := k1.Sign(data)
	if err != nil {
		t.Error("Sign failed:", err)
	}
	if !k2.Verify(data, s) {
		t.Error("Verify failed")
	}
}

func TestKeyFailed(t *testing.T) {
	if _, err := LoadPublicKey(nil); err != ErrInvalidKey {
		t.Error("Check load public key failed")
	}
	if _, err := LoadPublicKey([]byte("123")); err != ErrInvalidKey {
		t.Error("Check load invalid public key failed")
	}
	k, _ := GenerateSecretKey([]byte("123"))
	pk := k.GetPublicKey()
	pk.X.Set(big.NewInt(0))
	pk.Y.Set(big.NewInt(0))
	if _, err := pk.Encrypt(nil); err != nil {
		t.Error("Check encrypt data failed")
	}
	if _, err := k.Decrypt(nil); err != ErrInvalidMessage {
		t.Error("Check decrypt data failed")
	}
	if _, err := k.Decrypt([]byte{0x01}); err != ErrInvalidKey {
		t.Error("Check decrypt data failed")
	}
	if _, err := calcSharedKey(nil, nil); err != ErrInvalidKey {
		t.Error("Check calc shared key failed")
	}

}
