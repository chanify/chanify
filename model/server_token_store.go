package model

import (
	"database/sql"
	"errors"
)

// ServerTokenRecord stores metadata and a copy encrypted to the account public key.
// The raw bearer credential is never persisted in plaintext.
type ServerTokenRecord struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Expires        uint64 `json:"expires"`
	Created        int64  `json:"created"`
	Revoked        bool   `json:"revoked"`
	EncryptedToken string `json:"encrypted_token"`
	Hash           string `json:"-"`
}

type ServerTokenStore interface {
	CreateServerToken(uid string, record *ServerTokenRecord) error
	GetServerToken(uid, id string) (*ServerTokenRecord, error)
	ListServerTokens(uid string) ([]*ServerTokenRecord, error)
	UpdateServerToken(uid, id, name string, revoke bool) error
}

func createServerToken(db *sql.DB, uid string, r *ServerTokenRecord, mysql bool) error {
	query := "INSERT INTO `server_tokens` (`uid`,`id`,`name`,`expires`,`created`,`revoked`,`encrypted_token`,`hash`) VALUES(?,?,?,?,?,0,?,?) ON CONFLICT(`uid`,`id`) DO NOTHING"
	if mysql {
		query = "INSERT INTO `server_tokens` (`uid`,`id`,`name`,`expires`,`created`,`revoked`,`encrypted_token`,`hash`) VALUES(?,?,?,?,?,0,?,?) ON DUPLICATE KEY UPDATE `id`=`id`"
	}
	_, err := db.Exec(query, uid, r.ID, r.Name, r.Expires, r.Created, r.EncryptedToken, r.Hash)
	if err != nil {
		return err
	}
	existing, err := getServerToken(db, uid, r.ID)
	if err != nil {
		return err
	}
	if existing.Revoked || existing.Hash != r.Hash {
		return ErrInvalidToken
	}
	return nil
}
func getServerToken(db *sql.DB, uid, id string) (*ServerTokenRecord, error) {
	r := &ServerTokenRecord{}
	err := db.QueryRow("SELECT `id`,`name`,`expires`,`created`,`revoked`,`encrypted_token`,`hash` FROM `server_tokens` WHERE `uid`=? AND `id`=?", uid, id).Scan(&r.ID, &r.Name, &r.Expires, &r.Created, &r.Revoked, &r.EncryptedToken, &r.Hash)
	return r, err
}
func listServerTokens(db *sql.DB, uid string) ([]*ServerTokenRecord, error) {
	rows, err := db.Query("SELECT `id`,`name`,`expires`,`created`,`revoked`,`encrypted_token`,`hash` FROM `server_tokens` WHERE `uid`=? AND `revoked`=0 ORDER BY `created` DESC, `id` ASC", uid)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	records := []*ServerTokenRecord{}
	for rows.Next() {
		r := &ServerTokenRecord{}
		if err := rows.Scan(&r.ID, &r.Name, &r.Expires, &r.Created, &r.Revoked, &r.EncryptedToken, &r.Hash); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}
func updateServerToken(db *sql.DB, uid, id, name string, revoke bool) error {
	r, err := getServerToken(db, uid, id)
	if err != nil {
		return err
	}
	if revoke {
		_, err = db.Exec("UPDATE `server_tokens` SET `revoked`=1,`encrypted_token`='' WHERE `uid`=? AND `id`=?", uid, id)
		return err
	}
	if r.Revoked {
		return ErrInvalidToken
	}
	result, err := db.Exec("UPDATE `server_tokens` SET `name`=? WHERE `uid`=? AND `id`=? AND `revoked`=0", name, uid, id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 && r.Name != name {
		return errors.New("token no longer available")
	}
	return nil
}

func (s *sqlite) CreateServerToken(uid string, r *ServerTokenRecord) error {
	return createServerToken(s.db, uid, r, false)
}
func (s *mysql) CreateServerToken(uid string, r *ServerTokenRecord) error {
	return createServerToken(s.db, uid, r, true)
}
func (s *sqlite) GetServerToken(uid, id string) (*ServerTokenRecord, error) {
	return getServerToken(s.db, uid, id)
}
func (s *mysql) GetServerToken(uid, id string) (*ServerTokenRecord, error) {
	return getServerToken(s.db, uid, id)
}
func (s *sqlite) ListServerTokens(uid string) ([]*ServerTokenRecord, error) {
	return listServerTokens(s.db, uid)
}
func (s *mysql) ListServerTokens(uid string) ([]*ServerTokenRecord, error) {
	return listServerTokens(s.db, uid)
}
func (s *sqlite) UpdateServerToken(uid, id, name string, revoke bool) error {
	return updateServerToken(s.db, uid, id, name, revoke)
}
func (s *mysql) UpdateServerToken(uid, id, name string, revoke bool) error {
	return updateServerToken(s.db, uid, id, name, revoke)
}
