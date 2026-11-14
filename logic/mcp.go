package logic

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/chanify/chanify/crypto"
	"github.com/chanify/chanify/mcpnotify"
	"github.com/chanify/chanify/model"
)

// MCPChannelData converts a readable MCP channel code to the existing wire format.
func MCPChannelData(code string) ([]byte, error) { return mcpnotify.ChannelData(code) }
func (l *Logic) CreateMCPToken(uid, name string, expires int64) (*model.MCPTokenRecord, string, error) {
	store, ok := l.db.(model.MCPTokenStore)
	if !ok || l.srvless {
		return nil, "", model.ErrNotImplemented
	}
	user, err := l.GetUser(uid)
	if err != nil || user.IsServerless() {
		return nil, "", model.ErrInvalidUserID
	}
	name = strings.TrimSpace(name)
	if !utf8.ValidString(name) || name == "" || len(name) > 128 || expires <= time.Now().Unix() || expires > time.Now().AddDate(5, 0, 0).Unix() {
		return nil, "", errors.New("name and expiry (up to five years) are required")
	}
	id := make([]byte, 16)
	secret := make([]byte, 32)
	if _, err := rand.Read(id); err != nil {
		return nil, "", err
	}
	if _, err := rand.Read(secret); err != nil {
		return nil, "", err
	}
	raw := "cmcp1." + hex.EncodeToString(id) + "." + crypto.Base64Encode.EncodeToString(secret)
	r := &model.MCPTokenRecord{ID: hex.EncodeToString(id), UserID: uid, Name: name, Created: time.Now().Unix(), Expires: expires, Hash: serverTokenHash(raw)}
	if err := store.CreateMCPToken(r); err != nil {
		return nil, "", err
	}
	return r, raw, nil
}
func (l *Logic) AuthenticateMCP(raw string) (*model.MCPTokenRecord, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 3 || parts[0] != "cmcp1" || len(parts[1]) != 32 || len(parts[2]) != 43 {
		return nil, model.ErrInvalidToken
	}
	store, ok := l.db.(model.MCPTokenStore)
	if !ok || l.srvless {
		return nil, model.ErrInvalidToken
	}
	r, err := store.GetMCPToken(parts[1])
	if err != nil || r.Revoked || r.Expires <= time.Now().Unix() || subtle.ConstantTimeCompare([]byte(r.Hash), []byte(serverTokenHash(raw))) != 1 {
		return nil, model.ErrInvalidToken
	}
	u, err := l.GetUser(r.UserID)
	if err != nil || u.IsServerless() {
		return nil, model.ErrInvalidToken
	}
	return r, nil
}
func (l *Logic) ListMCPTokens(uid string) ([]*model.MCPTokenRecord, error) {
	s, ok := l.db.(model.MCPTokenStore)
	if !ok {
		return nil, model.ErrNotImplemented
	}
	return s.ListMCPTokens(uid)
}
func (l *Logic) UpdateMCPToken(uid, id, name string, revoke bool) error {
	s, ok := l.db.(model.MCPTokenStore)
	if !ok {
		return model.ErrNotImplemented
	}
	name = strings.TrimSpace(name)
	if !revoke && (name == "" || len(name) > 128) {
		return errors.New("invalid token name")
	}
	return s.UpdateMCPToken(uid, id, name, revoke)
}
