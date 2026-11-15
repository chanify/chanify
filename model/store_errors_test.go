package model

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func mockStore(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, m, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := m.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
		_ = db.Close()
	})
	return db, m
}

func TestStoreListFailures(t *testing.T) {
	for _, kind := range []string{"users", "tokens", "mcp", "sending"} {
		for _, mode := range []string{"query", "scan", "iteration"} {
			t.Run(kind+"/"+mode, func(t *testing.T) {
				db, m := mockStore(t)
				want := errors.New("database failure")
				var cols []string
				var values []interface{}
				switch kind {
				case "users":
					cols = []string{"uid", "devices"}
					values = []interface{}{"u", "invalid"}
				case "tokens":
					cols = []string{"kind", "id", "uid", "name", "expires", "created", "revoked"}
					values = []interface{}{"mcp", "id", "u", "name", "invalid", 1, false}
				case "mcp":
					cols = []string{"id", "uid", "name", "expires", "created", "revoked", "channels", "hash"}
					values = []interface{}{"id", "u", "name", "invalid", 1, false, "[]", "hash"}
				case "sending":
					cols = []string{"id", "name", "expires", "created", "revoked", "encrypted_token", "hash"}
					values = []interface{}{"id", "name", "invalid", 1, false, "cipher", "hash"}
				}
				q := m.ExpectQuery("SELECT")
				if mode == "query" {
					q.WillReturnError(want)
				} else {
					rows := sqlmock.NewRows(cols)
					// Driver values are concrete here so scan errors exercise conversion failures.
					for i := range values {
						if values[i] == "invalid" && mode == "iteration" {
							values[i] = int64(1)
						}
					}
					switch kind {
					case "users":
						rows.AddRow(values[0], values[1])
					case "tokens", "sending":
						rows.AddRow(values[0], values[1], values[2], values[3], values[4], values[5], values[6])
					default:
						rows.AddRow(values[0], values[1], values[2], values[3], values[4], values[5], values[6], values[7])
					}
					if mode == "iteration" {
						rows.RowError(0, want)
					}
					q.WillReturnRows(rows)
				}
				var err error
				switch kind {
				case "users":
					_, err = adminUsers(db)
				case "tokens":
					_, err = adminTokens(db)
				case "mcp":
					_, err = listMCPTokens(db, "u")
				case "sending":
					_, err = listServerTokens(db, "u")
				}
				if err == nil {
					t.Fatal("database failure ignored")
				}
				if mode != "scan" && !errors.Is(err, want) {
					t.Fatal(err)
				}
			})
		}
	}
}

func serverRows(revoked bool) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "name", "expires", "created", "revoked", "encrypted_token", "hash"}).AddRow("id", "name", 100, 1, revoked, "cipher", "hash")
}
func mcpRows(uid string, revoked bool) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "uid", "name", "expires", "created", "revoked", "channels", "hash"}).AddRow("id", uid, "name", 100, 1, revoked, "[]", "hash")
}

func TestServerTokenWriteFailures(t *testing.T) {
	for _, mode := range []string{"insert", "read", "hash", "revoked"} {
		t.Run("create/"+mode, func(t *testing.T) {
			db, m := mockStore(t)
			want := sql.ErrConnDone
			q := m.ExpectExec("INSERT INTO `server_tokens`")
			if mode == "insert" {
				q.WillReturnError(want)
			} else {
				q.WillReturnResult(sqlmock.NewResult(0, 1))
				r := m.ExpectQuery("SELECT")
				if mode == "read" {
					r.WillReturnError(want)
				} else {
					r.WillReturnRows(serverRows(mode == "revoked"))
					want = ErrInvalidToken
				}
			}
			hash := "hash"
			if mode == "hash" {
				hash = "different"
			}
			if err := createServerToken(db, "u", &ServerTokenRecord{ID: "id", Hash: hash}, false); !errors.Is(err, want) {
				t.Fatal(err)
			}
		})
	}
	for _, mode := range []string{"read", "revoked", "exec", "affected", "missing", "unchanged", "revoke"} {
		t.Run("update/"+mode, func(t *testing.T) {
			db, m := mockStore(t)
			want := sql.ErrConnDone
			q := m.ExpectQuery("SELECT")
			if mode == "read" {
				q.WillReturnError(want)
			} else {
				q.WillReturnRows(serverRows(mode == "revoked"))
				if mode == "revoked" {
					want = ErrInvalidToken
				} else {
					e := m.ExpectExec("UPDATE `server_tokens`")
					switch mode {
					case "exec", "revoke":
						e.WillReturnError(want)
					case "affected":
						e.WillReturnResult(sqlmock.NewErrorResult(want))
					default:
						e.WillReturnResult(sqlmock.NewResult(0, 0))
						want = nil
					}
				}
			}
			name := "new"
			if mode == "unchanged" {
				name = "name"
			}
			err := updateServerToken(db, "u", "id", name, mode == "revoke")
			if mode == "missing" {
				if err == nil || err.Error() != "token no longer available" {
					t.Fatal(err)
				}
			} else if !errors.Is(err, want) {
				t.Fatal(err)
			}
		})
	}
}

func TestMCPTokenOwnershipAndRevocation(t *testing.T) {
	for _, tc := range []struct {
		uid     string
		revoked bool
		want    error
	}{{"other", false, sql.ErrNoRows}, {"u", true, ErrInvalidToken}} {
		db, m := mockStore(t)
		m.ExpectQuery("SELECT").WillReturnRows(mcpRows(tc.uid, tc.revoked))
		if err := updateMCPToken(db, "u", "id", "new", false); !errors.Is(err, tc.want) {
			t.Fatal(err)
		}
	}
}

func TestMySQLTokenAndAdminAdapters(t *testing.T) {
	db, m := mockStore(t)
	s := &mysql{db: db}
	m.ExpectExec("INSERT INTO `mcp_tokens`").WithArgs("id", "u", "name", int64(100), int64(1), "[]", "hash").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := s.CreateMCPToken(&MCPTokenRecord{ID: "id", UserID: "u", Name: "name", Expires: 100, Created: 1, Hash: "hash"}); err != nil {
		t.Fatal(err)
	}
	m.ExpectQuery("SELECT").WithArgs("id").WillReturnRows(mcpRows("u", false))
	if r, err := s.GetMCPToken("id"); err != nil || r.Hash != "hash" {
		t.Fatal(r, err)
	}
	m.ExpectQuery("SELECT").WithArgs("u").WillReturnRows(mcpRows("u", false))
	if r, err := s.ListMCPTokens("u"); err != nil || len(r) != 1 {
		t.Fatal(r, err)
	}
	m.ExpectQuery("SELECT").WillReturnRows(mcpRows("u", false))
	m.ExpectExec("UPDATE `mcp_tokens`").WithArgs("new", "u", "id").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := s.UpdateMCPToken("u", "id", "new", false); err != nil {
		t.Fatal(err)
	}
	m.ExpectExec("INSERT INTO `server_tokens` .* ON DUPLICATE KEY UPDATE").WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectQuery("SELECT").WillReturnRows(serverRows(false))
	if err := s.CreateServerToken("u", &ServerTokenRecord{ID: "id", Hash: "hash"}); err != nil {
		t.Fatal(err)
	}
	m.ExpectQuery("SELECT").WillReturnRows(serverRows(false))
	if r, err := s.GetServerToken("u", "id"); err != nil || r.Hash != "hash" {
		t.Fatal(r, err)
	}
	m.ExpectQuery("SELECT").WillReturnRows(serverRows(false))
	if r, err := s.ListServerTokens("u"); err != nil || len(r) != 1 {
		t.Fatal(r, err)
	}
	m.ExpectQuery("SELECT").WillReturnRows(serverRows(false))
	m.ExpectExec("UPDATE `server_tokens`").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := s.UpdateServerToken("u", "id", "new", false); err != nil {
		t.Fatal(err)
	}
	m.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"uid", "devices"}).AddRow("u", 2))
	if r, err := s.AdminUsers(); err != nil || len(r) != 1 || r[0].Devices != 2 {
		t.Fatal(r, err)
	}
	m.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"kind", "id", "uid", "name", "expires", "created", "revoked"}).AddRow("sending", "id", "u", "name", 100, 1, false))
	if r, err := s.AdminTokens(); err != nil || len(r) != 1 || r[0].Kind != "sending" {
		t.Fatal(r, err)
	}
}

func TestDeleteUserTransactionFailures(t *testing.T) {
	for _, driver := range []string{"sqlite", "mysql"} {
		for fail := -1; fail <= 4; fail++ {
			t.Run(fmt.Sprintf("%s/failure-%d", driver, fail), func(t *testing.T) {
				db, m := mockStore(t)
				var store DB = &sqlite{db: db}
				if driver == "mysql" {
					store = &mysql{db: db}
				}
				want := sql.ErrConnDone
				if fail < 0 {
					m.ExpectBegin().WillReturnError(want)
				} else {
					m.ExpectBegin()
					tables := []string{"devices", "mcp_tokens", "server_tokens", "users"}
					for i, table := range tables {
						e := m.ExpectExec("DELETE FROM `" + table + "`").WithArgs("u")
						if i == fail {
							e.WillReturnError(want)
							break
						}
						e.WillReturnResult(sqlmock.NewResult(0, 1))
					}
					if fail < 4 {
						m.ExpectRollback()
					} else {
						m.ExpectCommit().WillReturnError(want)
					}
				}
				if err := store.DeleteUser("u"); !errors.Is(err, want) {
					t.Fatal(err)
				}
			})
		}
	}
	s := &nosql{}
	s.Close()
	if err := s.DeleteUser("u"); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLDeleteUserCommits(t *testing.T) {
	db, m := mockStore(t)
	m.ExpectBegin()
	for _, table := range []string{"devices", "mcp_tokens", "server_tokens", "users"} {
		m.ExpectExec("DELETE FROM `" + table + "`").WithArgs("u").WillReturnResult(sqlmock.NewResult(0, 1))
	}
	m.ExpectCommit()
	if err := (&mysql{db: db}).DeleteUser("u"); err != nil {
		t.Fatal(err)
	}
}
