package model

import (
	"database/sql"
	"log"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql" // mysql driver
)

type mysql struct {
	db *sql.DB
}

func init() {
	drivers["mysql"] = func(dsn string) (DB, error) {
		items := strings.Split(dsn, "://")
		db, _ := sql.Open(items[0], items[1])
		if db == nil {
			return nil, ErrInvalidDSN
		}
		log.Println("Open mysql database:", dsn)
		s := &mysql{db: db}
		if err := s.fixDB(); err != nil {
			return nil, err
		}
		return s, nil
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
	u := &User{UID: uid}
	row := s.db.QueryRow("SELECT `pubkey`, `seckey`, `flags` FROM `users` WHERE `uid`=? LIMIT 1;", uid)
	if err := row.Scan(&u.PublicKey, &u.SecretKey, &u.Flags); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *mysql) UpsertUser(u *User) error {
	_, err := s.db.Exec("INSERT INTO `users`(`uid`,`pubkey`,`seckey`,`flags`) VALUES(?,?,?,?) ON DUPLICATE KEY UPDATE `pubkey`=VALUES(`pubkey`),`seckey`=VALUES(`seckey`),`flags`=VALUES(`flags`);", u.UID, u.PublicKey, u.SecretKey, u.Flags)
	return err
}

func (s *mysql) BindDevice(uid string, uuid string, key []byte, devType int) error {
	_, err := s.db.Exec("INSERT INTO `devices`(`uuid`,`uid`,`key`,`type`) VALUES(?,?,?,?) ON DUPLICATE KEY UPDATE `uid`=VALUES(`uid`),`type`=VALUES(`type`);", uuid, uid, key, devType)
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
	rows, err := s.db.Query("SELECT `token`,`sandbox`,`type` FROM `devices` WHERE `uid`=? ORDER BY `lastupdate` DESC LIMIT 4;", uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		d := &Device{}
		rows.Scan(&d.Token, &d.Sandbox, &d.Type) // nolint: errcheck
		if len(d.Token) > 0 {
			devs = append(devs, d)
		}
	}
	return devs, nil
}

func (s *mysql) fixDB() error {
	s.db.SetConnMaxLifetime(time.Minute * 3)
	s.db.SetMaxOpenConns(10)
	s.db.SetMaxIdleConns(10)
	if err := s.db.Ping(); err != nil {
		return err
	}
	sqls := []string{
		"CREATE TABLE IF NOT EXISTS `options`(`key` VARCHAR(255), `value` VARBINARY(255), PRIMARY KEY (`key`));",
		"CREATE TABLE IF NOT EXISTS `users`(`uid` VARCHAR(255), `pubkey` VARBINARY(255) UNIQUE, `seckey` VARBINARY(255), `flags` INTEGER DEFAULT 0, `lastupdate` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, `createtime` TIMESTAMP DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY(`uid`));",
		"CREATE TABLE IF NOT EXISTS `devices`(`uuid` VARCHAR(255), `uid` VARCHAR(255), `key` VARBINARY(255), `type` INTEGER DEFAULT 0, `token` VARBINARY(255), `sandbox` INTEGER DEFAULT 0, `lastupdate` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, `createtime` TIMESTAMP DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY(`uuid`), INDEX(`uid`));",
	}
	for _, str := range sqls {
		if _, err := s.db.Exec(str); err != nil {
			return err
		}
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	cnt := 0
	row := tx.QueryRow("SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME='devices' AND COLUMN_NAME='type';")
	if err := row.Scan(&cnt); err != nil {
		tx.Rollback() // nolint: errcheck
		return err
	}
	if cnt <= 0 {
		if _, err := tx.Exec("ALTER TABLE `devices` ADD COLUMN `type` INTEGER DEFAULT 0 AFTER `key`;"); err != nil {
			tx.Rollback() // nolint: errcheck
			return err
		}
		log.Println("MySQL add column `type` into `devices`.")
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
