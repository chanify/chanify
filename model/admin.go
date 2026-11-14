package model

import "database/sql"

// Admin records deliberately exclude keys, bearer digests and encrypted copies.
type AdminUser struct {
	ID      string `json:"id"`
	Devices int    `json:"devices"`
}
type AdminToken struct {
	Kind    string `json:"kind"`
	ID      string `json:"id"`
	UserID  string `json:"user"`
	Name    string `json:"name"`
	Expires int64  `json:"expires"`
	Created int64  `json:"created"`
	Revoked bool   `json:"revoked"`
}
type AdminStore interface {
	AdminUsers() ([]AdminUser, error)
	AdminTokens() ([]AdminToken, error)
}

func adminUsers(db *sql.DB) ([]AdminUser, error) {
	rows, err := db.Query("SELECT `uid`,(SELECT COUNT(*) FROM `devices` WHERE `devices`.`uid`=`users`.`uid`) FROM `users` ORDER BY `uid`")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	result := []AdminUser{}
	for rows.Next() {
		var u AdminUser
		if err := rows.Scan(&u.ID, &u.Devices); err != nil {
			return nil, err
		}
		result = append(result, u)
	}
	return result, rows.Err()
}
func adminTokens(db *sql.DB) ([]AdminToken, error) {
	rows, err := db.Query("SELECT 'sending',`id`,`uid`,`name`,`expires`,`created`,`revoked` FROM `server_tokens` UNION ALL SELECT 'mcp',`id`,`uid`,`name`,`expires`,`created`,`revoked` FROM `mcp_tokens` ORDER BY `created` DESC,`id`")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	result := []AdminToken{}
	for rows.Next() {
		var r AdminToken
		if err := rows.Scan(&r.Kind, &r.ID, &r.UserID, &r.Name, &r.Expires, &r.Created, &r.Revoked); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, rows.Err()
}
func (s *sqlite) AdminUsers() ([]AdminUser, error)   { return adminUsers(s.db) }
func (s *mysql) AdminUsers() ([]AdminUser, error)    { return adminUsers(s.db) }
func (s *sqlite) AdminTokens() ([]AdminToken, error) { return adminTokens(s.db) }
func (s *mysql) AdminTokens() ([]AdminToken, error)  { return adminTokens(s.db) }
