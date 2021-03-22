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

	mock.ExpectExec("INSERT INTO `devices`").WillReturnResult(sqlmock.NewResult(1, 1))
	if err := db.BindDevice("123", "abc", []byte("key")); err != nil {
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

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS `options`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS `users`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS `devices`").WillReturnResult(sqlmock.NewResult(1, 1))
	if _, err := initDB(db.db, ""); err != nil {
		t.Fatal("Fix db failed:", err)
	}

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS `options`").WillReturnError(sql.ErrConnDone)
	if _, err := initDB(db.db, ""); err != sql.ErrConnDone {
		t.Fatal("Check fix db failed:", err)
	}

}

func TestMySQLPingFailed(t *testing.T) {
	db, mock, _ := sqlmock.New(sqlmock.MonitorPingsOption(true))
	defer db.Close()
	mock.ExpectPing().WillReturnError(sql.ErrConnDone)
	if _, err := initDB(db, ""); err != sql.ErrConnDone {
		t.Fatal("Check ping failed:", err)
	}

}

func TestMySQLFailed(t *testing.T) {
	if _, err := drivers["mysql"]("mysql://127.0.0.1:13306"); err == nil {
		t.Fatal("Cehck open mysql failed")
	}
}
