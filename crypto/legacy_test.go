package crypto

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"testing"
)

// These synthetic vectors pin the output of cc4bf1b (including edea261),
// whose go.mod declared Go 1.20. The P-256 scalar is the first valid 32-byte
// candidate from HKDF-Expand(SHA-1, secret, "chanify"), with no skipped byte.
// They were checked against cc4bf1b's source under Go 1.27.3 with the custom
// random reader compatibility setting enabled, not an old production binary.
func TestLegacySecretKeyVectors(t *testing.T) {
	vectors := []struct {
		secret, scalar, publicKey, nodeID string
	}{
		{
			"123",
			"ce36fc48135cc000bbd235fed7a7cabee7a22d30cee3122e2bb332635adbd1aa",
			"BGrLJKkPSTrfGHpfJbFF3qo_Lfl7FWE_ESBPFVMegjKQau1HE4S0gHkqBZHoinDBz_VAR9ZtrH2RT7RByJHv9K8",
			"AGIEVPVLCVE4HSEYWQC6GCSZYUWXNPTCDA",
		},
		{
			"123456",
			"94ea6fc53c3bc901b50b4a2598af7a76ff813126b3970d3633ed97bd03b13e14",
			"BOsr2fBwQzh7LxfWgopnp3pODfxsrZinPyD09TQAmr9x97JuMKCNJZ9c7FqvEVF43GjR6qSI7F0x2GGMnyNLvos",
			"AH2NOILMJBNZVWP7DEJGZHO52JFVVH2Q3E",
		},
		{
			"兼容性+ /?&=测试",
			"e1e1816c67fad2ff74650d5d1464704e03053810a4038e2eae1a1b27e2ae4b67",
			"BKfbN5ht4x0VQGuYObg2eI6B3UlN6IPkqjimpRLe8oWo7qnM1ij7dcfIAWjf-_M1FoCP-Yh73Uj2eMbKtumpv5s",
			"AGXEONADDA77VMW2EBVISF4EGECBITMPBE",
		},
	}
	for _, customRand := range []string{"0", "1"} {
		t.Run("cryptocustomrand="+customRand, func(t *testing.T) {
			t.Setenv("GODEBUG", "cryptocustomrand="+customRand)
			for i, v := range vectors {
				t.Run(fmt.Sprint(i), func(t *testing.T) {
					key := GenerateSecretKey([]byte(v.secret))
					scalar, err := key.Bytes()
					if err != nil {
						t.Fatal(err)
					}
					if hex.EncodeToString(scalar) != v.scalar {
						t.Error("legacy private scalar changed")
					}
					if key.EncodePublicKey() != v.publicKey {
						t.Error("legacy public key changed")
					}
					if key.ToID(1) != v.nodeID {
						t.Error("legacy node ID changed")
					}
				})
			}
		})
	}
}

func TestLegacyStoredKeyAndMessages(t *testing.T) {
	// Synthetic PEM, signature, and ECIES message produced by cc4bf1b's source
	// with secret "123" and message "legacy compatibility".
	const pemKey = `-----BEGIN PRIVATE KEY-----
MHcCAQEEIM42/EgTXMAAu9I1/tenyr7noi0wzuMSLiuzMmNa29GqoAoGCCqGSM49
AwEHoUQDQgAEasskqQ9JOt8Yel8lsUXeqj8t+XsVYT8RIE8VUx6CMpBq7UcThLSA
eSoFkeiKcMHP9UBH1m2sfZFPtEHIke/0rw==
-----END PRIVATE KEY-----
`
	decode := func(s string) []byte {
		t.Helper()
		data, err := hex.DecodeString(s)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	stored, err := LoadSecretKey([]byte(pemKey))
	if err != nil {
		t.Fatal(err)
	}
	derived := GenerateSecretKey([]byte("123"))
	if !stored.Equal(&derived.PrivateKey) {
		t.Fatal("derived key differs from stored legacy key")
	}
	if !bytes.Equal(derived.MarshalSecretKey(), []byte(pemKey)) {
		t.Error("legacy PEM encoding changed")
	}
	message := []byte("legacy compatibility")
	signature := decode("304402204b282356759c6e8c49eb3d22d8b88fa2f9e9d6830ee276367defd998a083c5670220277db51a2b3f79bb131301ae442563fadbc93c1e23490a4240a079954ebda900")
	if !derived.GetPublicKey().Verify(message, signature) {
		t.Error("legacy signature rejected")
	}
	if derived.GetPublicKey().Verify([]byte("tampered message"), signature) {
		t.Error("tampered message accepted")
	}
	badSignature := append([]byte(nil), signature...)
	badSignature[len(badSignature)-1] ^= 1
	if derived.GetPublicKey().Verify(message, badSignature) {
		t.Error("tampered signature accepted")
	}
	if derived.GetPublicKey().Verify(message, []byte{0}) {
		t.Error("malformed signature accepted")
	}
	ciphertext := decode("04a6b9db0a1d4c53ca508a4a027b9bf10339883c03dd4a688e3b087c4323403cad7d15abea4ed55390629ed032863d7f93f413bb9f58819dd92a4114718b5081b936a49660cf889cb21cbf4d6556ca3fe852df5fc44f5a720eb9c0d346c139eb80aae3c853")
	plaintext, err := derived.Decrypt(ciphertext)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(plaintext, message) {
		t.Error("legacy ciphertext decrypted incorrectly")
	}
}

// TestLegacyClientSignature uses the historical fixture in ccsrv/model/utils_test.go,
// committed before this upgrade, rather than a newly generated signature.
func TestLegacyClientSignature(t *testing.T) {
	decode := func(s string) []byte {
		t.Helper()
		data, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	key, err := LoadPublicKey(decode("BBc64c0GiltQB86I5TPH6/1Z6qf1q/8vU0EyR0HWr15m3B0efSgSWBm1dX63SvwUhTYe6j9XivGT/ZYgZDq6jI4="))
	if err != nil {
		t.Fatal(err)
	}
	signature := decode("MEYCIQDZ7gzkd3iSP/T3LfBp11RjJyISdGwEgJpd611jI7z2vgIhAOoGuG1iFXioCSSAUSfbu1+sUOkOjqb0P3XsLpMIfZ01")
	if !key.Verify([]byte("hello world!"), signature) {
		t.Fatal("historical client signature rejected")
	}
	if key.Verify([]byte("tampered"), signature) {
		t.Fatal("tampered message accepted")
	}
}
