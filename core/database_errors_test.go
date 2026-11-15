package core

import (
	"database/sql"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chanify/chanify/crypto"
	"github.com/chanify/chanify/logic"
)

func persistentCore(t *testing.T) (*Core, *sql.DB) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "node.db")
	c := New()
	if err := c.Init(&logic.Options{DBUrl: "sqlite://" + path, Registerable: true}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(c.Close)
	key := crypto.GenerateSecretKey([]byte("synthetic-mcp-account"))
	if _, err := c.logic.UpsertUser(key.ToID(0), key.EncodePublicKey(), false); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return c, db
}

func TestManagementDatabaseFailures(t *testing.T) {
	for _, tc := range []struct {
		path, table string
		body        map[string]interface{}
	}{
		{"/rest/v1/mcp-tokens", "mcp_tokens", map[string]interface{}{"action": "list"}},
		{"/rest/v1/server-tokens", "server_tokens", map[string]interface{}{"action": "list"}},
		{"/rest/v1/delete-user", "devices", map[string]interface{}{"key": crypto.GenerateSecretKey([]byte("synthetic-mcp-account")).EncodePublicKey()}},
	} {
		t.Run(tc.table, func(t *testing.T) {
			c, db := persistentCore(t)
			if _, err := db.Exec("DROP TABLE `" + tc.table + "`"); err != nil {
				t.Fatal(err)
			}
			w := signedManagement(t, c, tc.path, tc.body)
			if w.Code != 500 {
				t.Fatal(w.Code, w.Body.String())
			}
		})
	}
}

func TestAdminSnapshotDatabaseFailure(t *testing.T) {
	c, db := persistentCore(t)
	if _, err := db.Exec("DROP TABLE `users`"); err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	c.adminSnapshot(w, httptest.NewRequest("GET", "/", nil))
	if w.Code != 500 {
		t.Fatal(w.Code, w.Body.String())
	}
}

func TestAdminInventoryStorageFailure(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "images"), []byte("not a directory"), 0600); err != nil {
		t.Fatal(err)
	}
	c := New()
	if err := c.Init(&logic.Options{DBUrl: "sqlite://?mode=memory", FilePath: root}); err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	w := httptest.NewRecorder()
	c.adminMediaList(w, httptest.NewRequest("GET", "/", nil))
	if w.Code != 500 || !strings.Contains(w.Body.String(), "error") {
		t.Fatal(w.Code, w.Body.String())
	}
}
