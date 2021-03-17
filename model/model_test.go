package model

import (
	"net/url"
	"testing"
)

func TestInitDB(t *testing.T) {
	drivers["mock"] = func(dsn *url.URL) (DB, error) {
		return nil, nil
	}
	_, err := InitDB("mock://")
	if err != nil {
		t.Fatal("Open database failed")
	}
}

func TestInitDBFailed(t *testing.T) {
	if _, err := InitDB("nodb://"); err != ErrDriverNotFound {
		t.Fatal("Check init db driver failed!")
	}
	if _, err := InitDB("::"); err == nil {
		t.Fatal("Check invalid dsn failed!")
	}
}
