package model

import (
	"io/ioutil"
	"net/url"
	"os"
	"testing"
)

func TestSqliteOpen(t *testing.T) {
	file, err := ioutil.TempFile("", "db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())
	u, _ := url.Parse("sqlite://" + file.Name())
	db, err := drivers["sqlite"](u)
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
	if _, err := db.GetUser("123"); err == nil {
		t.Fatal("Check not found user failed")
	}
	usr := &User{Uid: "abc", Flags: 123}
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

func TestSqliteOpenFailed(t *testing.T) {
	u := &url.URL{}
	open := drivers["sqlite"]
	u.Path = ".."
	if _, err := open(u); err == nil {
		t.Fatal("Check sqlite connect failed")
	}

	u.Path = "?mode=ro&vfs=unix"
	if _, err := open(u); err == nil {
		t.Fatal("Check sqlite fix failed")
	}
}
