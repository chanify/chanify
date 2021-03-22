package model

import (
	"database/sql"
	"log"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type mysql struct {
	db *sql.DB
}

func init() {
	drivers["mysql"] = func(dsn string) (DB, error) {
		items := strings.Split(dsn, "://")
		db, _ := sql.Open(items[0], items[1])
		return initDB(db, dsn)
	}
}

func (s *mysql) Close() {
	if s.db != nil {
		s.db.Close()
		s.db = nil
		log.Println("Close mysql database")
	}
}

func (s *mysql) GetOption(key string, value interface{}) error {
	row := s.db.QueryRow("SELECT `value` FROM `options` WHERE `key`=? LIMIT 1;", key)
	return row.Scan(value)
}

func (s *mysql) SetOption(key string, value interface{}) error {
	_, err := s.db.Exec("INSERT INTO `options`(`key`,`value`) VALUES(?,?) ON DUPLICATE KEY UPDATE `value`=VALUES(`value`);", key, value)
	return err
}

func (s *mysql) GetUser(uid string) (*User, error) {
	u := &User{Uid: uid}
	row := s.db.QueryRow("SELECT `pubkey`, `seckey`, `flags` FROM `users` WHERE `uid`=? LIMIT 1;", uid)
	if err := row.Scan(&u.PublicKey, &u.SecretKey, &u.Flags); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *mysql) UpsertUser(u *User) error {
	_, err := s.db.Exec("INSERT INTO `users`(`uid`,`pubkey`,`seckey`,`flags`) VALUES(?,?,?,?) ON DUPLICATE KEY UPDATE `pubkey`=VALUES(`pubkey`),`seckey`=VALUES(`seckey`),`flags`=VALUES(`flags`);", u.Uid, u.PublicKey, u.SecretKey, u.Flags)
	return err
}

func (s *mysql) BindDevice(uid string, uuid string, key []byte) error {
	_, err := s.db.Exec("INSERT INTO `devices`(`uuid`,`uid`,`key`) VALUES(?,?,?) ON DUPLICATE KEY UPDATE `uid`=VALUES(`uid`);", uuid, uid, key)
	return err

}

func (s *mysql) UnbindDevice(uid string, uuid string) error {
	_, err := s.db.Exec("DELETE FROM `devices` WHERE `uuid`=? AND `uid`=?;", uuid, uid)
	return err
}

func (s *mysql) UpdatePushToken(uid string, uuid string, token []byte, sandbox bool) error {
	_, err := s.db.Exec("UPDATE `devices` SET `uid`=?,`token`=?,`sandbox`=? WHERE `uuid`=?;", uid, token, sandbox, uuid)
	return err
}

func (s *mysql) GetDeviceKey(uuid string) ([]byte, error) {
	var key []byte
	row := s.db.QueryRow("SELECT `key` FROM `devices` WHERE `uuid`=? LIMIT 1;", uuid)
	err := row.Scan(&key)
	return key, err
}

func (s *mysql) GetDevices(uid string) ([]*Device, error) {
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

func initDB(db *sql.DB, dsn string) (DB, error) {
	s := &mysql{db: db}
	if s.db == nil {
		return nil, ErrInvalidDSN
	}
	s.db.SetConnMaxLifetime(time.Minute * 3)
	s.db.SetMaxOpenConns(10)
	s.db.SetMaxIdleConns(10)
	if err := s.db.Ping(); err != nil {
		return nil, err
	}
	sqls := []string{
		"CREATE TABLE IF NOT EXISTS `options`(`key` VARCHAR(255), `value` VARBINARY(255), PRIMARY KEY (`key`));",
		"CREATE TABLE IF NOT EXISTS `users`(`uid` VARCHAR(255), `pubkey` VARBINARY(255) UNIQUE, `seckey` VARBINARY(255), `flags` INTEGER DEFAULT 0, `lastupdate` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, `createtime` TIMESTAMP DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY(`uid`));",
		"CREATE TABLE IF NOT EXISTS `devices`(`uuid` VARCHAR(255), `uid` VARCHAR(255) , `key` VARBINARY(255), `token` VARBINARY(255), `sandbox` INTEGER DEFAULT 0, `lastupdate` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, `createtime` TIMESTAMP DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY(`uuid`), INDEX(`uid`));",
	}
	for _, str := range sqls {
		if _, err := s.db.Exec(str); err != nil {
			return nil, err
		}
	}
	log.Println("Open mysql database:", dsn)
	return s, nil
}
