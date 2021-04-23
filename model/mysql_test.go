package model

import (
	"database/sql"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestMySQL(t *testing.T) {
	dbmock, mock, _ := sqlmock.New()
	db := &mysql{db: dbmock}
	defer db.Close()

	mock.ExpectExec("INSERT INTO `options`").WillReturnResult(sqlmock.NewResult(1, 1))
	if err := db.SetOption("secret", "hello"); err != nil {
		t.Fatal("Set option failed:", err)
	}

	mock.ExpectQuery("SELECT `value` FROM `options`").WillReturnRows(sqlmock.NewRows([]string{""}).AddRow("123456"))
	var secret string
	if err := db.GetOption("secret", &secret); err != nil || secret != "123456" {
		t.Fatal("Set option failed:", err)
	}

	mock.ExpectQuery("SELECT `pubkey`, `seckey`, `flags` FROM `users`").
		WillReturnRows(sqlmock.NewRows([]string{"", "", ""}).AddRow([]byte("123"), []byte("abc"), 100))
	if _, err := db.GetUser("123"); err != nil {
		t.Fatal("Get user failed:", err)
	}

	mock.ExpectQuery("SELECT `pubkey`, `seckey`, `flags` FROM `users`").WillReturnError(sql.ErrNoRows)
	if _, err := db.GetUser("123"); err != sql.ErrNoRows {
		t.Fatal("Check get user failed:", err)
	}

	mock.ExpectExec("INSERT INTO `users`").WillReturnResult(sqlmock.NewResult(1, 1))
	if err := db.UpsertUser(&User{}); err != nil {
		t.Fatal("Upsert user failed:", err)
	}
}

func TestMySQLDevice(t *testing.T) {
	dbmock, mock, _ := sqlmock.New()
	db := &mysql{db: dbmock}
	defer db.Close()

	mock.ExpectExec("INSERT INTO `devices`").WillReturnResult(sqlmock.NewResult(1, 1))
	if err := db.BindDevice("123", "abc", []byte("key"), 0); err != nil {
		t.Fatal("Bind device failed:", err)
	}

	mock.ExpectExec("DELETE FROM `devices`").WillReturnResult(sqlmock.NewResult(1, 1))
	if err := db.UnbindDevice("123", "abc"); err != nil {
		t.Fatal("Bind device failed:", err)
	}

	mock.ExpectExec("UPDATE `devices`").WillReturnResult(sqlmock.NewResult(1, 1))
	if err := db.UpdatePushToken("123", "abc", []byte("token"), false); err != nil {
		t.Fatal("Bind device failed:", err)
	}

	mock.ExpectQuery("SELECT `key` FROM `devices`").WillReturnRows(sqlmock.NewRows([]string{"key"}).AddRow("123456"))
	k, err := db.GetDeviceKey("123")
	if err != nil || string(k) != "123456" {
		t.Fatal("Get device key failed:", err)
	}

	mock.ExpectQuery("SELECT `token`,`sandbox` FROM `devices`").WillReturnRows(sqlmock.NewRows([]string{"token", "sandbox"}).AddRow("123", true))
	if _, err := db.GetDevices("1"); err != nil {
		t.Fatal("Get devices failed:", err)
	}

	mock.ExpectQuery("SELECT `token`,`sandbox` FROM `devices`").WillReturnError(sql.ErrNoRows)
	if _, err := db.GetDevices("1"); err != sql.ErrNoRows {
		t.Fatal("Check get devices failed")
	}
}

func TestMySQLFixDB(t *testing.T) {
	dbmock, mock, _ := sqlmock.New()
	db := &mysql{db: dbmock}
	defer db.Close()

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS `options`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS `users`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS `devices`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT COUNT(.+) FROM INFORMATION_SCHEMA.COLUMNS").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectCommit()
	if err := db.fixDB(); err != nil {
		t.Fatal("Fix db failed:", err)
	}

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS `options`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS `users`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS `devices`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT COUNT(.+) FROM INFORMATION_SCHEMA.COLUMNS").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec("ALTER TABLE `devices` ADD COLUMN `type` ").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	if err := db.fixDB(); err != nil {
		t.Fatal("Fix db failed:", err)
	}
}

func TestMySQLFixDBFailed(t *testing.T) {
	dbmock, mock, _ := sqlmock.New()
	db := &mysql{db: dbmock}
	defer db.Close()

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS `options`").WillReturnError(sql.ErrConnDone)
	if err := db.fixDB(); err != sql.ErrConnDone {
		t.Fatal("Check fix db failed:", err)
	}

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS `options`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS `users`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS `devices`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectBegin().WillReturnError(sql.ErrConnDone)
	if err := db.fixDB(); err != sql.ErrConnDone {
		t.Fatal("Check fix db begin failed:", err)
	}

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS `options`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS `users`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS `devices`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT COUNT(.+) FROM INFORMATION_SCHEMA.COLUMNS").WillReturnError(sql.ErrConnDone)
	if err := db.fixDB(); err != sql.ErrConnDone {
		t.Fatal("Check fix db select column failed:", err)
	}

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS `options`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS `users`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS `devices`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT COUNT(.+) FROM INFORMATION_SCHEMA.COLUMNS").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec("ALTER TABLE `devices` ADD COLUMN `type` ").WillReturnError(sql.ErrConnDone)
	if err := db.fixDB(); err != sql.ErrConnDone {
		t.Fatal("Check fix db add column failed:", err)
	}

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS `options`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS `users`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS `devices`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT COUNT(.+) FROM INFORMATION_SCHEMA.COLUMNS").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectCommit().WillReturnError(sql.ErrConnDone)
	if err := db.fixDB(); err != sql.ErrConnDone {
		t.Fatal("Check fix db commit failed:", err)
	}
}

func TestMySQLPingFailed(t *testing.T) {
	dbmock, mock, _ := sqlmock.New(sqlmock.MonitorPingsOption(true))
	db := &mysql{db: dbmock}
	defer db.Close()
	mock.ExpectPing().WillReturnError(sql.ErrConnDone)
	if err := db.fixDB(); err != sql.ErrConnDone {
		t.Fatal("Check ping failed:", err)
	}
}

func TestMySQLFailed(t *testing.T) {
	if _, err := drivers["mysql"]("mysql://127.0.0.1:13306"); err == nil {
		t.Fatal("Check open mysql failed")
	}
}

func TestMySQLOpenFailed(t *testing.T) {
	open := drivers["mysql"]
	dbmock, mock, _ := sqlmock.NewWithDSN("sqlmock")
	defer dbmock.Close()
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS `options`").WillReturnError(sql.ErrConnDone)
	if _, err := open("sqlmock://sqlmock"); err != sql.ErrConnDone {
		t.Error("Check open mysql failed:", err)
	}

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS `options`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS `users`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS `devices`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT COUNT(.+) FROM INFORMATION_SCHEMA.COLUMNS").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectCommit()
	if _, err := open("sqlmock://sqlmock"); err != nil {
		t.Error("Open mysql driver failed:", err)
	}
}
