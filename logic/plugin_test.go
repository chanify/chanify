package logic

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestLoadLuaFile(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "plugin")
	defer os.RemoveAll(dir)
	whdir := filepath.Join(dir, "webhook")
	os.MkdirAll(whdir, os.ModePerm) // nolint: errcheck
	fpath := filepath.Join(whdir, "github.lua")
	fs, err := os.Create(fpath)
	if err != nil {
		t.Fatal(err)
	}
	defer fs.Close()
	fs.WriteString("-") // nolint: errcheck
	fs.Sync()           // nolint: errcheck

	opts := []map[string]interface{}{
		{
			"name": "github",
			"file": "webhook/github.lua",
		},
	}
	l := loadWebhookPlugin(dir, opts)
	defer l.Close()
	if l.loadLuaFile("not_exist", dir) != nil {
		t.Error("Check not exist lua file failed")
	}
	l.ReloadWebhook(fpath)
	fs.WriteString("-") // nolint: errcheck
	fs.Sync()           // nolint: errcheck
	l.ReloadWebhook(fpath)
}

func TestLuaWatch(t *testing.T) {
	opts := []map[string]interface{}{}
	l := loadWebhookPlugin("", opts)
	defer l.Close()
	l.watcher.Errors <- errors.New("123")
}

func TestWebHookDoCall(t *testing.T) {
	l := lua.NewState()
	defer l.Close()
	w := &Webhook{}
	if w.DoCall(l) != ErrNotFound {
		t.Error("Check do call failed")
	}
	w = &Webhook{
		lfunc: &luaFunc{},
	}
	if w.DoCall(l) != ErrNotFound {
		t.Error("Check do call failed")
	}
}
