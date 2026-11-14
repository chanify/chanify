package logic

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/chanify/chanify/crypto"
	"github.com/chanify/chanify/model"
)

func serverTokenHash(raw string) string {
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}
func (l *Logic) ManageServerTokens(uid, action, id, name, raw string) ([]*model.ServerTokenRecord, error) {
	store, ok := l.db.(model.ServerTokenStore)
	if !ok || l.srvless {
		return nil, model.ErrNotImplemented
	}
	user, err := l.GetUser(uid)
	if err != nil || user.IsServerless() {
		return nil, model.ErrInvalidUserID
	}
	name = strings.TrimSpace(name)
	if action == "create" || action == "rename" {
		if name == "" || !utf8.ValidString(name) || len(name) > 128 {
			return nil, errors.New("name must contain 1 to 128 UTF-8 bytes")
		}
	}
	switch action {
	case "list":
	case "create":
		tk, parseErr := model.ParseToken(raw)
		if parseErr != nil || !tk.IsServerToken() || tk.GetUserID() != uid || !l.verifyTokenSignature(tk) {
			return nil, model.ErrInvalidToken
		}
		encrypted := user.PublicKeyEncrypt([]byte(raw))
		if len(encrypted) == 0 {
			return nil, model.ErrInvalidToken
		}
		err = store.CreateServerToken(uid, &model.ServerTokenRecord{ID: tk.ServerTokenID(), Name: name, Expires: tk.ExpiresAt(), Created: time.Now().Unix(), Hash: serverTokenHash(raw), EncryptedToken: crypto.Base64Encode.EncodeToString(encrypted)})
	case "rename", "revoke":
		decoded, decodeErr := hex.DecodeString(id)
		if decodeErr != nil || len(decoded) != 16 {
			return nil, model.ErrInvalidToken
		}
		err = store.UpdateServerToken(uid, id, name, action == "revoke")
	default:
		return nil, model.ErrNotImplemented
	}
	if err != nil {
		return nil, err
	}
	return store.ListServerTokens(uid)
}
