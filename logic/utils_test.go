package logic

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFixPath(t *testing.T) {
	fpath := filepath.Join(os.TempDir(), "files")
	defer os.RemoveAll(fpath)
	if err := fixPath(fpath + "/tests"); err != nil {
		t.Error("Fix path failed")
	}
}

func TestSaveFile(t *testing.T) {
	if err := saveFile("/", nil); err != nil {
		t.Error("Check save failed failed", err)
	}
}

func TestCompileLua(t *testing.T) {
	fpath := filepath.Join(os.TempDir(), "compile_lua_test")
	fs, err := os.Create(fpath)
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(fpath)
	defer fs.Close()
	if _, err := compileLua(fpath); err != nil {
		t.Fatal("Compile lua failed")
	}
	fs.WriteString("break") // nolint: errcheck
	fs.Sync()               // nolint: errcheck
	if _, err := compileLua(fpath); err == nil {
		t.Fatal("Check compile lua failed", err)
	}
}

func TestCompileLuaFailed(t *testing.T) {
	fpath := filepath.Join(os.TempDir(), "compile_lua_not_exist")
	if _, err := compileLua(fpath); err == nil {
		t.Error("Check compile lua path failed")
	}
	if _, err := compileLua("/"); err == nil {
		t.Error("Check compile lua chunck failed")
	}
}
