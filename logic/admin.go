package logic

import (
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/chanify/chanify/model"
)

func (l *Logic) AdminAvailable() bool { _, ok := l.db.(model.AdminStore); return ok && !l.srvless }
func (l *Logic) AdminRecords() ([]model.AdminUser, []model.AdminToken, error) {
	store, ok := l.db.(model.AdminStore)
	if !ok {
		return nil, nil, model.ErrNotImplemented
	}
	users, err := store.AdminUsers()
	if err != nil {
		return nil, nil, err
	}
	tokens, err := store.AdminTokens()
	return users, tokens, err
}
func (l *Logic) AdminUpdateToken(kind, uid, id, name string, revoke bool) error {
	switch kind {
	case "mcp":
		return l.UpdateMCPToken(uid, id, name, revoke)
	case "sending":
		store, ok := l.db.(model.ServerTokenStore)
		if !ok {
			return model.ErrNotImplemented
		}
		name = strings.TrimSpace(name)
		if !revoke && (name == "" || len(name) > 128 || !utf8.ValidString(name)) {
			return errors.New("invalid name")
		}
		return store.UpdateServerToken(uid, id, name, revoke)
	default:
		return errors.New("invalid credential type")
	}
}
