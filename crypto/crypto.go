package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"hash"

	"golang.org/x/crypto/hkdf"
)

type PublicKey struct {
	ecdsa.PublicKey
}

type SecretKey struct {
	ecdsa.PrivateKey
}

var (
	eciesKeyLen       = aes.BlockSize
	ErrInvalidKey     = errors.New("InvalidKey")
	ErrInvalidMessage = errors.New("InvalidMessage")
)

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

func GenerateSecretKey(secret []byte) (*SecretKey, error) {
	r := hkdf.Expand(sha1.New, secret, []byte("chanify"))
	key, _ := ecdsa.GenerateKey(elliptic.P256(), r)
	k := &SecretKey{}
	k.PrivateKey = *key
	return k, nil
}

func (k *PublicKey) MarshalPublicKey() []byte {
	return marshalPublicKey(&k.PublicKey)
}

func (k *PublicKey) ToID(code byte) string {
	return formatToID(code, k.MarshalPublicKey())
}

func (k *PublicKey) Verify(msg []byte, sig []byte) bool {
	hash := sha256.Sum256([]byte(msg))
	return ecdsa.VerifyASN1(&k.PublicKey, hash[:], sig)
}

func (k *PublicKey) Encrypt(data []byte) ([]byte, error) {
	seckey, _ := ecdsa.GenerateKey(k.Curve, rand.Reader)
	Z, _ := calcSharedKey(seckey, &k.PublicKey)
	X := elliptic.Marshal(k.Curve, seckey.X, seckey.Y)
	K := X963KDF(Z, X, sha256.New)
	key := K[:eciesKeyLen]
	nonce := K[eciesKeyLen:]
	block, _ := aes.NewCipher(key)
	aesgcm, _ := cipher.NewGCMWithNonceSize(block, eciesKeyLen)
	cipherData := aesgcm.Seal(nil, nonce, data, nil)
	return append(X, cipherData...), nil
}

func (k *SecretKey) GetPublicKey() *PublicKey {
	key := &PublicKey{}
	key.PublicKey = k.PublicKey
	return key
}

func (k *SecretKey) MarshalPublicKey() []byte {
	return marshalPublicKey(&k.PublicKey)
}

func (k *SecretKey) ToID(code byte) string {
	return formatToID(code, k.MarshalPublicKey())
}

func (k *SecretKey) EncodePublicKey() string {
	return base64.RawStdEncoding.EncodeToString(k.MarshalPublicKey())
}

func (k *SecretKey) Sign(msg []byte) ([]byte, error) {
	hash := sha256.Sum256([]byte(msg))
	return ecdsa.SignASN1(rand.Reader, &k.PrivateKey, hash[:])
}

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
	K := X963KDF(Z, X, sha256.New)
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
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(data)
}

func X963KDF(sharedKeySeed []byte, ephemeralPublicKey []byte, hfnc func() hash.Hash) []byte {
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
		counter += 1
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
