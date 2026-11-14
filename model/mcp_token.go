package model

import (
	"database/sql"
)

type MCPTokenRecord struct {
	ID      string `json:"id"`
	UserID  string `json:"user"`
	Name    string `json:"name"`
	Expires int64  `json:"expires"`
	Created int64  `json:"created"`
	Revoked bool   `json:"revoked"`
	Hash    string `json:"-"`
}
type MCPTokenStore interface {
	CreateMCPToken(*MCPTokenRecord) error
	GetMCPToken(string) (*MCPTokenRecord, error)
	ListMCPTokens(string) ([]*MCPTokenRecord, error)
	UpdateMCPToken(string, string, string, bool) error
}

func createMCPToken(db *sql.DB, r *MCPTokenRecord) error {
	// Keep the obsolete column for existing database compatibility; it grants no scope.
	_, err := db.Exec("INSERT INTO `mcp_tokens` (`id`,`uid`,`name`,`expires`,`created`,`revoked`,`channels`,`hash`) VALUES(?,?,?,?,?,0,?,?)", r.ID, r.UserID, r.Name, r.Expires, r.Created, "[]", r.Hash)
	return err
}
func scanMCPToken(row interface{ Scan(...interface{}) error }) (*MCPTokenRecord, error) {
	r := &MCPTokenRecord{}
	var channels string
	err := row.Scan(&r.ID, &r.UserID, &r.Name, &r.Expires, &r.Created, &r.Revoked, &channels, &r.Hash)

	return r, err
}
func getMCPToken(db *sql.DB, id string) (*MCPTokenRecord, error) {
	return scanMCPToken(db.QueryRow("SELECT `id`,`uid`,`name`,`expires`,`created`,`revoked`,`channels`,`hash` FROM `mcp_tokens` WHERE `id`=?", id))
}
func listMCPTokens(db *sql.DB, uid string) ([]*MCPTokenRecord, error) {
	rows, err := db.Query("SELECT `id`,`uid`,`name`,`expires`,`created`,`revoked`,`channels`,`hash` FROM `mcp_tokens` WHERE `uid`=? ORDER BY `created` DESC,`id`", uid)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	records := []*MCPTokenRecord{}
	for rows.Next() {
		r, err := scanMCPToken(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}
func updateMCPToken(db *sql.DB, uid, id, name string, revoke bool) error {
	r, err := getMCPToken(db, id)
	if err != nil {
		return err
	}
	if r.UserID != uid {
		return sql.ErrNoRows
	}
	if revoke {
		_, err = db.Exec("UPDATE `mcp_tokens` SET `revoked`=1 WHERE `uid`=? AND `id`=?", uid, id)
		return err
	}
	if r.Revoked {
		return ErrInvalidToken
	}
	_, err = db.Exec("UPDATE `mcp_tokens` SET `name`=? WHERE `uid`=? AND `id`=? AND `revoked`=0", name, uid, id)
	return err
}
func (s *sqlite) CreateMCPToken(r *MCPTokenRecord) error         { return createMCPToken(s.db, r) }
func (s *mysql) CreateMCPToken(r *MCPTokenRecord) error          { return createMCPToken(s.db, r) }
func (s *sqlite) GetMCPToken(id string) (*MCPTokenRecord, error) { return getMCPToken(s.db, id) }
func (s *mysql) GetMCPToken(id string) (*MCPTokenRecord, error)  { return getMCPToken(s.db, id) }
func (s *sqlite) ListMCPTokens(uid string) ([]*MCPTokenRecord, error) {
	return listMCPTokens(s.db, uid)
}
func (s *mysql) ListMCPTokens(uid string) ([]*MCPTokenRecord, error) { return listMCPTokens(s.db, uid) }
func (s *sqlite) UpdateMCPToken(uid, id, name string, revoke bool) error {
	return updateMCPToken(s.db, uid, id, name, revoke)
}
func (s *mysql) UpdateMCPToken(uid, id, name string, revoke bool) error {
	return updateMCPToken(s.db, uid, id, name, revoke)
}
