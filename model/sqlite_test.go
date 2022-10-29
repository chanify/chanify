package model

import (
	"database/sql"
	"os"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestSqliteOpen(t *testing.T) {
	file, err := os.CreateTemp("", "db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())
	db, err := drivers["sqlite"]("sqlite://" + file.Name())
	if err != nil {
		t.Fatal("Open sqlite failed")
	}
	defer db.Close()
	if err := db.SetOption("version", 123); err != nil {
		t.Fatal("Set option failed")
	}
	var ver int
	if err := db.GetOption("version", &ver); err != nil || ver != 123 {
		t.Fatal("Get option failed")
	}
	if err := db.GetOption("name", &ver); err == nil {
		t.Fatal("Get option failed")
	}
}

func TestSqliteUser(t *testing.T) {
	db, _ := drivers["sqlite"]("sqlite://?mode=memory")
	defer db.Close()
	if _, err := db.GetUser("123"); err == nil {
		t.Fatal("Check not found user failed")
	}
	usr := &User{UID: "abc", Flags: 123}
	if err := db.UpsertUser(usr); err != nil {
		t.Fatal("Upsert user failed:", err)
	}
	uu, err := db.GetUser("abc")
	if err != nil {
		t.Fatal("Get user failed:", err)
	}
	if uu.Flags != usr.Flags {
		t.Fatal("Store user failed:", err)
	}
	usr.Flags = 456
	if err := db.UpsertUser(usr); err != nil {
		t.Fatal("Update user failed:", err)
	}
	uu, err = db.GetUser("abc")
	if err != nil {
		t.Fatal("Get user again failed:", err)
	}
	if uu.Flags != usr.Flags {
		t.Fatal("Overwrite user failed:", err)
	}
}

func TestSqliteDevice(t *testing.T) {
	db, _ := drivers["sqlite"]("sqlite://?mode=memory")
	defer db.Close()
	if err := db.BindDevice("abc", "xyz", []byte("key"), 0); err != nil {
		t.Fatal("Bind device failed:", err)
	}
	if err := db.UpdatePushToken("abc", "xyz", []byte("PushToken"), false); err != nil {
		t.Fatal("Update push token failed:", err)
	}
	devs, err := db.GetDevices("abc")
	if err != nil {
		t.Fatal("Get devices failed:", err)
	}
	if len(devs) != 1 || string(devs[0].Token) != "PushToken" {
		t.Fatal("Get push token failed")
	}
	if err := db.UnbindDevice("abc", "xyz"); err != nil {
		t.Fatal("Unbind device failed:", err)
	}
}

func TestSqliteGetDeviceFailed(t *testing.T) {
	db, _ := drivers["sqlite"]("sqlite://?mode=memory")
	defer db.Close()
	db.(*sqlite).db.Exec("DROP TABLE `devices`;") // nolint: errcheck
	if _, err := db.GetDevices("123"); err == nil {
		t.Error("Check get devices failed")
	}
}

func TestSqliteOpenFailed(t *testing.T) {
	open := drivers["sqlite"]
	if _, err := open("sqlite:///?mode=readonly"); err == nil {
		t.Fatal("Check sqlite connect failed")
	}
	if _, err := open("sqlite://?mode=ro&vfs=unix"); err == nil {
		t.Fatal("Check sqlite fix failed")
	}
}

func TestSQLiteFixDB(t *testing.T) {
	dbmock, mock, _ := sqlmock.New()
	db := &sqlite{db: dbmock}
	defer db.Close()

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS `options`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT COUNT(.+) FROM pragma_table_info").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec("ALTER TABLE `devices` ADD COLUMN `type`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	if err := db.fixDB(); err != nil {
		t.Fatal("Fix db failed:", err)
	}
}

func TestSQLiteFixDBFailed(t *testing.T) {
	dbmock, mock, _ := sqlmock.New()
	db := &sqlite{db: dbmock}
	defer db.Close()

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS `options`").WillReturnError(sql.ErrConnDone)
	if err := db.fixDB(); err != sql.ErrConnDone {
		t.Fatal("Check fix db failed:", err)
	}

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS `options`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectBegin().WillReturnError(sql.ErrConnDone)
	if err := db.fixDB(); err != sql.ErrConnDone {
		t.Fatal("Check fix db begin failed:", err)
	}

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS `options`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT COUNT(.+) FROM pragma_table_info").WillReturnError(sql.ErrConnDone)
	if err := db.fixDB(); err != sql.ErrConnDone {
		t.Fatal("Check fix db select column failed:", err)
	}

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS `options`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT COUNT(.+) FROM pragma_table_info").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec("ALTER TABLE `devices` ADD COLUMN `type`").WillReturnError(sql.ErrConnDone)
	if err := db.fixDB(); err != sql.ErrConnDone {
		t.Fatal("Check fix db add column failed:", err)
	}

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS `options`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT COUNT(.+) FROM pragma_table_info").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectCommit().WillReturnError(sql.ErrConnDone)
	if err := db.fixDB(); err != sql.ErrConnDone {
		t.Fatal("Check fix db commit failed:", err)
	}
}
