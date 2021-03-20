package model

import (
	"database/sql"
	"log"
	"net/url"
	"strings"

	_ "modernc.org/sqlite"
)

type sqlite struct {
	db *sql.DB
}

func init() {
	drivers["sqlite"] = func(dsn *url.URL) (DB, error) {
		db, _ := sql.Open("sqlite", "file:"+dsn.Path)
		if err := db.Ping(); err != nil {
			return nil, err
		}
		log.Println("Open sqlite database:", dsn.Path)
		s := &sqlite{db: db}
		if err := s.fixDB(); err != nil {
			return nil, err
		}
		return s, nil
	}
}

func (s *sqlite) Close() {
	if s.db != nil {
		s.db.Close()
		s.db = nil
		log.Println("Close sqlite database")
	}
}

func (s *sqlite) SetOption(key string, value interface{}) error {
	_, err := s.db.Exec("INSERT INTO `options`(`key`,`value`) VALUES(?,?) ON CONFLICT(`key`) DO UPDATE SET `value`=excluded.`value`;", key, value)
	return err
}

func (s *sqlite) GetOption(key string, value interface{}) error {
	row := s.db.QueryRow("SELECT `value` FROM `options` WHERE `key`=? LIMIT 1;", key)
	return row.Scan(value)
}

func (s *sqlite) GetUser(uid string) (*User, error) {
	u := &User{Uid: uid}
	row := s.db.QueryRow("SELECT `pubkey`, `seckey`, `flags` FROM `users` WHERE `uid`=? LIMIT 1;", uid)
	if err := row.Scan(&u.PublicKey, &u.SecretKey, &u.Flags); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *sqlite) UpsertUser(u *User) error {
	_, err := s.db.Exec("INSERT INTO `users`(`uid`,`pubkey`,`seckey`,`flags`) VALUES(?,?,?,?) ON CONFLICT(`uid`) DO UPDATE SET `pubkey`=excluded.`pubkey`,`seckey`=excluded.`seckey`,`flags`=excluded.`flags`,`lastupdate`=CURRENT_TIMESTAMP;", u.Uid, u.PublicKey, u.SecretKey, u.Flags)
	return err
}

func (s *sqlite) BindDevice(uid string, uuid string, key []byte) error {
	_, err := s.db.Exec("INSERT INTO `devices`(`uuid`,`uid`,`key`) VALUES(?,?,?) ON CONFLICT(`uuid`) DO UPDATE SET `uid`=excluded.`uid`,`lastupdate`=CURRENT_TIMESTAMP;", uuid, uid, key)
	return err
}

func (s *sqlite) UnbindDevice(uid string, uuid string) error {
	_, err := s.db.Exec("DELETE FROM `devices` WHERE `uuid`=? AND `uid`=?;", uuid, uid)
	return err
}

func (s *sqlite) UpdatePushToken(uid string, uuid string, token []byte, sandbox bool) error {
	_, err := s.db.Exec("UPDATE `devices` SET `uid`=?,`token`=?,`sandbox`=?,`lastupdate`=CURRENT_TIMESTAMP WHERE `uuid`=?;", uid, token, sandbox, uuid)
	return err
}

func (s *sqlite) GetDeviceKey(uuid string) ([]byte, error) {
	var key []byte
	row := s.db.QueryRow("SELECT `key` FROM `devices` WHERE `uuid`=? LIMIT 1;", uuid)
	err := row.Scan(&key)
	return key, err
}

func (s *sqlite) GetDevices(uid string) ([]*Device, error) {
	devs := []*Device{}
	rows, err := s.db.Query("SELECT `token`,`sandbox` FROM `devices` WHERE `uid`=? ORDER BY `lastupdate` DESC LIMIT 4;", uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		d := &Device{}
		rows.Scan(&d.Token, &d.Sandbox) // nolint: errcheck
		if len(d.Token) > 0 {
			devs = append(devs, d)
		}
	}
	return devs, nil
}

func (s *sqlite) fixDB() error {
	sqls := []string{
		"CREATE TABLE IF NOT EXISTS `options`(`key` TEXT PRIMARY KEY, `value` BLOB);",
		"CREATE TABLE IF NOT EXISTS `users`(`uid` TEXT PRIMARY KEY, `pubkey` BLOB UNIQUE, `seckey` BLOB, `flags` INTEGER DEFAULT 0, `lastupdate` TIMESTAMP DEFAULT CURRENT_TIMESTAMP, `createtime` TIMESTAMP DEFAULT CURRENT_TIMESTAMP);",
		"CREATE TABLE IF NOT EXISTS `devices`(`uuid` TEXT PRIMARY KEY, `uid` TEXT, `key` BLOB, `token` BLOB, `sandbox` INTEGER DEFAULT 0, `lastupdate` TIMESTAMP DEFAULT CURRENT_TIMESTAMP, `createtime` TIMESTAMP DEFAULT CURRENT_TIMESTAMP);",
		"CREATE INDEX IF NOT EXISTS `idx_devices_uid` ON `devices`(`uid`);",
	}
	_, err := s.db.Exec(strings.Join(sqls, ""))
	return err
}
