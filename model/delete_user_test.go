package model

import (
	"database/sql"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestDeleteUserPreservesOtherAccounts(t *testing.T) {
	db, err := InitDB("sqlite://?mode=memory")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, uid := range []string{"delete", "keep"} {
		if err := db.UpsertUser(&User{UID: uid, PublicKey: []byte(uid)}); err != nil {
			t.Fatal(err)
		}
		if err := db.BindDevice(uid, uid+"device", []byte(uid), 0); err != nil {
			t.Fatal(err)
		}
		if err := db.UpdatePushToken(uid, uid+"device", []byte("test-token"), false); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.DeleteUser("delete"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.GetUser("delete"); err != sql.ErrNoRows {
		t.Fatalf("deleted account remains: %v", err)
	}
	if devices, err := db.GetDevices("delete"); err != nil || len(devices) != 0 {
		t.Fatalf("deleted devices remain: %v", err)
	}
	if _, err := db.GetDeviceKey("deletedevice"); err != sql.ErrNoRows {
		t.Fatalf("deleted device key remains: %v", err)
	}
	if _, err := db.GetUser("keep"); err != nil {
		t.Fatalf("other account lost: %v", err)
	}
	if devices, err := db.GetDevices("keep"); err != nil || len(devices) != 1 {
		t.Fatalf("other devices lost: %v", err)
	}
	if key, err := db.GetDeviceKey("keepdevice"); err != nil || string(key) != "keep" {
		t.Fatalf("other device key changed: %v", err)
	}
	if err := db.DeleteUser("delete"); err != nil {
		t.Fatalf("retry failed: %v", err)
	}
}

func TestDeleteUserRollsBackOnFailure(t *testing.T) {
	for _, driver := range []string{"sqlite", "mysql"} {
		t.Run(driver, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				mock.ExpectClose()
				if err := db.Close(); err != nil {
					t.Error(err)
				}
			}()
			mock.ExpectBegin()
			mock.ExpectExec("DELETE FROM `devices`").WithArgs("uid").WillReturnResult(sqlmock.NewResult(0, 2))
			mock.ExpectExec("DELETE FROM `mcp_tokens`").WithArgs("uid").WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectExec("DELETE FROM `server_tokens`").WithArgs("uid").WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectExec("DELETE FROM `users`").WithArgs("uid").WillReturnError(sql.ErrConnDone)
			mock.ExpectRollback()
			var store DB = &sqlite{db: db}
			if driver == "mysql" {
				store = &mysql{db: db}
			}
			if err := store.DeleteUser("uid"); err != sql.ErrConnDone {
				t.Fatalf("lost failure: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
