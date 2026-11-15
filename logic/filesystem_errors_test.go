package logic

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chanify/chanify/model"
)

func writeTestFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
}
func denyAccess(t *testing.T, path string) {
	t.Helper()
	if os.Geteuid() == 0 {
		t.Skip("permission checks require an unprivileged user")
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chmod(path, info.Mode().Perm()); err != nil {
			t.Error(err)
		}
	})
}

func TestMediaUnreadableFiles(t *testing.T) {
	l := mediaFixture(t)
	path, err := l.SaveFile("files", []byte("synthetic"), "test.txt")
	if err != nil {
		t.Fatal(err)
	}
	id := strings.TrimPrefix(path, "/files/")
	denyAccess(t, l.metadataPath(id))
	if _, err := l.mediaMetadata(id); !os.IsPermission(err) {
		t.Fatalf("want permission failure: %v", err)
	}
	denyAccess(t, filepath.Join(l.filepath, id))
	if _, _, err := l.OpenMedia(id); !os.IsPermission(err) {
		t.Fatalf("want permission failure: %v", err)
	}
}

func TestMediaUnreadableDirectories(t *testing.T) {
	for _, mode := range []string{"list", "lookup", "save", "delete"} {
		t.Run(mode, func(t *testing.T) {
			l := mediaFixture(t)
			if mode == "lookup" {
				denyAccess(t, l.filepath)
			} else {
				denyAccess(t, filepath.Join(l.filepath, "images"))
			}
			id := "images/" + strings.Repeat("a", 40)
			switch mode {
			case "list":
				if _, err := l.MediaInventory(); !os.IsPermission(err) {
					t.Fatal(err)
				}
			case "lookup":
				if _, err := l.mediaPath(id); !os.IsPermission(err) {
					t.Fatal(err)
				}
			case "save":
				if _, err := l.SaveFile("images", []byte("test")); !os.IsPermission(err) {
					t.Fatal(err)
				}
			case "delete":
				if _, err := l.DeleteMedia(nil, "snapshot"); !os.IsPermission(err) {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestSaveMediaMetadataFailures(t *testing.T) {
	for _, mode := range []string{"directory", "temp", "rename"} {
		t.Run(mode, func(t *testing.T) {
			l := mediaFixture(t)
			id := "files/" + strings.Repeat("a", 40)
			dir := filepath.Join(l.filepath, ".metadata")
			if mode == "directory" {
				writeTestFile(t, dir, nil)
			} else {
				if err := os.Mkdir(dir, 0700); err != nil {
					t.Fatal(err)
				}
				if mode == "temp" {
					denyAccess(t, dir)
				} else {
					if err := os.Mkdir(l.metadataPath(id), 0700); err != nil {
						t.Fatal(err)
					}
				}
			}
			if err := l.saveMediaMetadata(id, []byte("data"), nil); err == nil {
				t.Fatal("metadata failure ignored")
			}
		})
	}
}

func TestMediaMetadataKindsAndAliasLimit(t *testing.T) {
	l := mediaFixture(t)
	for _, tc := range []struct {
		data []byte
		kind string
	}{
		{append([]byte("RIFF\x00\x00\x00\x00WAVEfmt "), make([]byte, 32)...), "audio"},
		{append([]byte{0, 0, 0, 24, 'f', 't', 'y', 'p', 'm', 'p', '4', '2'}, make([]byte, 12)...), "video"},
	} {
		id := "files/" + strings.Repeat("a", 40)
		if err := l.saveMediaMetadata(id, tc.data, []string{"."}); err != nil {
			t.Fatal(err)
		}
		m, err := l.mediaMetadata(id)
		if err != nil || m.Kind != tc.kind {
			t.Fatal(m, err)
		}
	}
	names := make([]string, 1025)
	for i := range names {
		names[i] = strings.Repeat("x", i%250+1) + string(rune(0x4e00+i))
	}
	id := "files/" + strings.Repeat("b", 40)
	if err := l.saveMediaMetadata(id, []byte("data"), names); err != nil {
		t.Fatal(err)
	}
	m, err := l.mediaMetadata(id)
	if err != nil || len(m.Names) != 1024 {
		t.Fatal(len(m.Names), err)
	}
	if mediaName(".") != "" {
		t.Fatal("dot name retained")
	}
}

func TestSaveFilePreservesExistingContents(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file")
	if err := saveFile(path, []byte("original")); err != nil {
		t.Fatal(err)
	}
	if err := saveFile(path, []byte("replacement")); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(data, []byte("original")) {
		t.Fatal(string(data), err)
	}
	if err := saveFile(filepath.Join(path, "child"), nil); err == nil {
		t.Fatal("invalid destination accepted")
	}
	if err := saveMediaBytes(filepath.Join(path, "child"), nil); err == nil {
		t.Fatal("invalid destination accepted")
	}
}

func TestLoadDBRejectsMalformedSecret(t *testing.T) {
	path := filepath.Join(t.TempDir(), "node.db")
	dsn := "sqlite://" + path
	db, err := model.InitDB(dsn)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.SetOption("secret", []byte("not a key")); err != nil {
		t.Fatal(err)
	}
	db.Close()
	if _, err := NewLogic(&Options{DBUrl: dsn}); err == nil {
		t.Fatal("malformed persisted secret accepted")
	}
	if _, err := NewLogic(&Options{Secret: "synthetic", MCPEnabled: true}); err == nil {
		t.Fatal("stateless MCP accepted")
	}
	l := mediaFixture(t)
	if _, err := l.SaveFile("unknown", []byte("data")); !errors.Is(err, ErrInvalidContent) {
		t.Fatal(err)
	}
}
