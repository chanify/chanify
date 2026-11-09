package logic

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	lua "github.com/yuin/gopher-lua"
)

func TestLoadLuaFile(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "plugin")
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Error(err)
		}
	}()
	whdir := filepath.Join(dir, "webhook")
	os.MkdirAll(whdir, os.ModePerm) // nolint: errcheck
	fpath := filepath.Join(whdir, "github.lua")
	fs, err := os.Create(fpath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := fs.Close(); err != nil {
			t.Error(err)
		}
	}()
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
	// Own the channels in this test; fsnotify owns and closes its real channels.
	watcher := &fsnotify.Watcher{
		Events: make(chan fsnotify.Event),
		Errors: make(chan error),
	}
	l := &pluginManager{}
	done := make(chan struct{})
	go func() {
		l.luaWatch(watcher)
		close(done)
	}()
	watcher.Errors <- errors.New("123")
	close(watcher.Errors)
	close(watcher.Events)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("watch loop did not stop after channels closed")
	}
}

func TestPluginCloseWaitsForWatcher(t *testing.T) {
	l := loadWebhookPlugin("", nil)
	l.Close()
	// Closing twice must remain safe after the watch goroutine exits.
	l.Close()
	if l.watcher != nil {
		t.Fatal("watcher remains after close")
	}
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
