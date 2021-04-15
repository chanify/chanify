package logic

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFixPath(t *testing.T) {
	fpath := filepath.Join(os.TempDir(), "files")
	defer os.RemoveAll(fpath)
	if err := FixPath(fpath + "/tests"); err != nil {
		t.Error("Fix path failed")
	}
}

func TestSaveFile(t *testing.T) {
	if err := SaveFile("/", nil); err != nil {
		t.Error("Check save failed failed", err)
	}
}
