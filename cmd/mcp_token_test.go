//go:build !test

package cmd

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/chanify/chanify/crypto"
	"github.com/chanify/chanify/logic"
	"github.com/chanify/chanify/model"
	"github.com/spf13/viper"
)

func TestMCPTokenCommandsLifecycle(t *testing.T) {
	resetConfig(t)
	dsn := "sqlite://" + filepath.Join(t.TempDir(), "node.db")
	opts := &logic.Options{DBUrl: dsn, Registerable: true}
	l, err := logic.NewLogic(opts)
	if err != nil {
		t.Fatal(err)
	}
	key := crypto.GenerateSecretKey([]byte("synthetic-cli-management"))
	uid := key.ToID(0)
	if _, err := l.UpsertUser(uid, key.EncodePublicKey(), false); err != nil {
		l.Close()
		t.Fatal(err)
	}
	l.Close()
	viper.Set("server.dburl", dsn)
	invoke := func(action string, flags map[string]string) ([]byte, error) {
		t.Helper()
		c := commandForTest(t, "mcp-token", action)
		setFlag(t, c, "user", uid)
		for name, value := range flags {
			setFlag(t, c, name, value)
		}
		var out bytes.Buffer
		c.SetOut(&out)
		err := c.RunE(c, nil)
		return out.Bytes(), err
	}
	data, err := invoke("create", map[string]string{"name": "Agent", "expires-in": "1h"})
	if err != nil {
		t.Fatal(err)
	}
	var created struct {
		Credential model.MCPTokenRecord `json:"credential"`
		Token      string               `json:"token"`
	}
	if err := json.Unmarshal(data, &created); err != nil {
		t.Fatal(err)
	}
	if created.Credential.ID == "" || created.Token == "" || created.Credential.Name != "Agent" {
		t.Fatal("missing credential metadata")
	}
	if _, err := invoke("rename", map[string]string{"id": created.Credential.ID, "name": "Renamed"}); err != nil {
		t.Fatal(err)
	}
	data, err = invoke("list", map[string]string{"dburl": dsn})
	if err != nil {
		t.Fatal(err)
	}
	var list []model.MCPTokenRecord
	if err := json.Unmarshal(data, &list); err != nil || len(list) != 1 || list[0].Name != "Renamed" {
		t.Fatal(list, err)
	}
	if bytes.Contains(data, []byte(created.Token)) {
		t.Fatal("list leaked raw credential")
	}
	if _, err := invoke("revoke", map[string]string{"id": created.Credential.ID}); err != nil {
		t.Fatal(err)
	}
	l, err = logic.NewLogic(opts)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	if _, err := l.AuthenticateMCP(created.Token); err == nil {
		t.Fatal("CLI revocation not persisted")
	}
	if _, err := invoke("rename", map[string]string{"id": created.Credential.ID, "name": "Resurrect"}); err == nil {
		t.Fatal("revoked credential renamed")
	}
}

func TestMCPTokenCommandFailures(t *testing.T) {
	resetConfig(t)
	for _, tc := range []struct{ action, user, dsn, ttl string }{
		{"create", "", "", "1h"},
		{"list", "user", "missing-driver://", ""},
		{"create", "missing", "sqlite://?mode=memory", "0s"},
		{"create", "missing", "sqlite://?mode=memory", "100000h"},
		{"create", "missing", "sqlite://?mode=memory", "1h"},
		{"rename", "missing", "sqlite://?mode=memory", ""},
	} {
		t.Run(tc.action+"/"+tc.ttl, func(t *testing.T) {
			c := commandForTest(t, "mcp-token", tc.action)
			setFlag(t, c, "user", tc.user)
			setFlag(t, c, "dburl", tc.dsn)
			if tc.ttl != "" {
				setFlag(t, c, "expires-in", tc.ttl)
			}
			if err := c.RunE(c, nil); err == nil {
				t.Fatal("invalid management request accepted")
			}
		})
	}
	c := commandForTest(t, "mcp-token", "list")
	setFlag(t, c, "user", "user")
	setFlag(t, c, "dburl", "sqlite://?mode=memory")
	c.SetOut(outputFailure{})
	if err := c.RunE(c, nil); err == nil {
		t.Fatal("output failure ignored")
	}
}
