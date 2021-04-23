package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"encoding/pem"
	"errors"
	"hash"
	"io"

	"golang.org/x/crypto/hkdf"
)

// variable define
var (
	eciesKeyLen       = aes.BlockSize
	Base64Encode      = base64.RawURLEncoding
	Base32Encode      = base32.StdEncoding.WithPadding(base32.NoPadding)
	ErrInvalidKey     = errors.New("InvalidKey")
	ErrInvalidMessage = errors.New("InvalidMessage")
)

// PublicKey of ECDSA
type PublicKey struct {
	ecdsa.PublicKey
}

// SecretKey of ECDSA
type SecretKey struct {
	ecdsa.PrivateKey
}

// LoadPublicKey from binary data
func LoadPublicKey(key []byte) (*PublicKey, error) {
	pk := &PublicKey{}
	pk.Curve = elliptic.P256()
	if len(key) <= 0 {
		return nil, ErrInvalidKey
	}
	pk.X, pk.Y = elliptic.Unmarshal(pk.Curve, key)
	if pk.X == nil || pk.Y == nil {
		return nil, ErrInvalidKey
	}
	return pk, nil
}

// LoadSecretKey from binary data
func LoadSecretKey(key []byte) (*SecretKey, error) {
	block, _ := pem.Decode(key)
	if block == nil {
		return nil, ErrInvalidKey
	}
	sk := &SecretKey{}
	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	sk.PrivateKey = *privateKey
	return sk, nil
}

// GenerateSecretKey with secret key
func GenerateSecretKey(secret []byte) *SecretKey {
	var r io.Reader
	if len(secret) > 0 {
		r = hkdf.Expand(sha1.New, secret, []byte("chanify"))
	} else {
		r = rand.Reader
	}
	key, _ := ecdsa.GenerateKey(elliptic.P256(), r)
	k := &SecretKey{}
	k.PrivateKey = *key
	return k
}

// MarshalPublicKey return binary public key
func (k *PublicKey) MarshalPublicKey() []byte {
	return marshalPublicKey(&k.PublicKey)
}

// ToID calc public key id with code
func (k *PublicKey) ToID(code byte) string {
	return formatToID(code, k.MarshalPublicKey())
}

// Verify message with sign
func (k *PublicKey) Verify(msg []byte, sig []byte) bool {
	hash := sha256.Sum256([]byte(msg))
	return ecdsa.VerifyASN1(&k.PublicKey, hash[:], sig)
}

// Encrypt data with public key
func (k *PublicKey) Encrypt(data []byte) ([]byte, error) {
	seckey, _ := ecdsa.GenerateKey(k.Curve, rand.Reader)
	Z, _ := calcSharedKey(seckey, &k.PublicKey)
	X := elliptic.Marshal(k.Curve, seckey.X, seckey.Y)
	K := x963KDF(Z, X, sha256.New)
	key := K[:eciesKeyLen]
	nonce := K[eciesKeyLen:]
	block, _ := aes.NewCipher(key)
	aesgcm, _ := cipher.NewGCMWithNonceSize(block, eciesKeyLen)
	cipherData := aesgcm.Seal(nil, nonce, data, nil)
	return append(X, cipherData...), nil
}

// GetPublicKey from secret key
func (k *SecretKey) GetPublicKey() *PublicKey {
	key := &PublicKey{}
	key.PublicKey = k.PublicKey
	return key
}

// MarshalSecretKey return binary secret key
func (k *SecretKey) MarshalSecretKey() []byte {
	encoded, _ := x509.MarshalECPrivateKey(&k.PrivateKey)
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: encoded})
}

// MarshalPublicKey return binary public key
func (k *SecretKey) MarshalPublicKey() []byte {
	return marshalPublicKey(&k.PublicKey)
}

// ToID calc secret key id with code
func (k *SecretKey) ToID(code byte) string {
	return formatToID(code, k.MarshalPublicKey())
}

// EncodePublicKey return base64 public key from secret key
func (k *SecretKey) EncodePublicKey() string {
	return Base64Encode.EncodeToString(k.MarshalPublicKey())
}

// Sign message with secret key
func (k *SecretKey) Sign(msg []byte) ([]byte, error) {
	hash := sha256.Sum256([]byte(msg))
	return ecdsa.SignASN1(rand.Reader, &k.PrivateKey, hash[:])
}

// Decrypt data with sercet key
func (k *SecretKey) Decrypt(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, ErrInvalidMessage
	}
	var rLen int
	switch data[0] {
	case 2, 3, 4:
		rLen = (k.PublicKey.Curve.Params().BitSize + 7) / 4
	default:
		return nil, ErrInvalidKey
	}

	R := ecdsa.PublicKey{}
	R.Curve = elliptic.P256()
	R.X, R.Y = elliptic.Unmarshal(R.Curve, data[:rLen])

	Z, _ := calcSharedKey(&k.PrivateKey, &R)

	X := data[:rLen]
	K := x963KDF(Z, X, sha256.New)
	key := K[:eciesKeyLen]
	nonce := K[eciesKeyLen:]
	block, _ := aes.NewCipher(key)
	aesgcm, _ := cipher.NewGCMWithNonceSize(block, eciesKeyLen)
	return aesgcm.Open(nil, nonce, data[rLen:], nil)
}

func marshalPublicKey(key *ecdsa.PublicKey) []byte {
	return elliptic.Marshal(key.Curve, key.X, key.Y)
}

func formatToID(code byte, key []byte) string {
	s1 := sha256.Sum256(key)
	var data []byte
	data = append(data, s1[:]...)
	data = append(data, key...)
	s2 := sha1.Sum(data)
	data = append([]byte{code}, s2[:]...)
	return Base32Encode.EncodeToString(data)
}

func x963KDF(sharedKeySeed []byte, ephemeralPublicKey []byte, hfnc func() hash.Hash) []byte {
	length := 32
	output := make([]byte, 0)
	outlen := 0
	counter := uint32(1)

	for outlen < length {
		h := hfnc()
		h.Write(sharedKeySeed) // nolint: errcheck

		counterBuf := make([]byte, 4)
		binary.BigEndian.PutUint32(counterBuf, counter)
		h.Write(counterBuf) // nolint: errcheck

		h.Write(ephemeralPublicKey) // nolint: errcheck

		out := h.Sum(nil)
		output = append(output, out...)
		outlen += h.Size()
		counter++
	}
	return output
}

func calcSharedKey(sec *ecdsa.PrivateKey, pub *ecdsa.PublicKey) ([]byte, error) {
	if pub == nil {
		return nil, ErrInvalidKey
	}
	res, _ := sec.Curve.ScalarMult(pub.X, pub.Y, sec.D.Bytes())
	byteLen := sec.Params().BitSize / 8
	key := res.Bytes()
	if byteLen > len(key) {
		fixKey := make([]byte, byteLen)
		copy(fixKey[byteLen-len(key):], key)
		key = fixKey
	}
	return key, nil
}
