package logic

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chanify/chanify/model"
)

type failingFile struct {
	*os.File
	writeErr, closeErr error
}

func (f *failingFile) Write(data []byte) (int, error) {
	if f.writeErr != nil {
		return 0, f.writeErr
	}
	return f.File.Write(data)
}
func (f *failingFile) Close() error {
	err := f.File.Close()
	if f.closeErr != nil {
		return f.closeErr
	}
	return err
}

func TestAtomicAttachmentWriteFailures(t *testing.T) {
	for _, stage := range []string{"stat", "create", "write", "close", "rename"} {
		t.Run(stage, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "attachment")
			ops := systemFiles
			want := errors.New("injected " + stage)
			switch stage {
			case "stat":
				ops.Stat = func(string) (fs.FileInfo, error) { return nil, want }
			case "create":
				ops.CreateTemp = func(string, string) (writableFile, error) { return nil, want }
			case "rename":
				ops.Rename = func(string, string) error { return want }
			default:
				ops.CreateTemp = func(dir, pattern string) (writableFile, error) {
					f, err := os.CreateTemp(dir, pattern)
					if err != nil {
						return nil, err
					}
					wrapped := &failingFile{File: f}
					if stage == "write" {
						wrapped.writeErr = want
					} else {
						wrapped.closeErr = want
					}
					return wrapped, nil
				}
			}
			if err := saveMediaBytesWithFiles(&ops, path, []byte("complete data")); !errors.Is(err, want) {
				t.Fatal(err)
			}
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Fatalf("failed upload published: %v", err)
			}
			entries, err := os.ReadDir(dir)
			if err != nil || len(entries) != 0 {
				t.Fatalf("temporary upload leaked: %v %v", entries, err)
			}
		})
	}
}

func TestAtomicMetadataWriteFailures(t *testing.T) {
	for _, stage := range []string{"directory-race", "write", "close", "rename"} {
		t.Run(stage, func(t *testing.T) {
			l := mediaFixture(t)
			ops := systemFiles
			l.fileOps = &ops
			want := errors.New("injected " + stage)
			switch stage {
			case "directory-race":
				calls := 0
				ops.Lstat = func(path string) (fs.FileInfo, error) {
					if filepath.Base(path) == ".metadata" {
						calls++
						if calls == 2 {
							return nil, want
						}
					}
					return os.Lstat(path)
				}
			case "rename":
				ops.Rename = func(string, string) error { return want }
			default:
				ops.CreateTemp = func(dir, pattern string) (writableFile, error) {
					f, err := os.CreateTemp(dir, pattern)
					if err != nil {
						return nil, err
					}
					wrapped := &failingFile{File: f}
					if stage == "write" {
						wrapped.writeErr = want
					} else {
						wrapped.closeErr = want
					}
					return wrapped, nil
				}
			}
			id := "files/" + strings.Repeat("a", 40)
			err := l.saveMediaMetadata(id, []byte("complete data"), nil)
			if stage == "directory-race" {
				if !errors.Is(err, ErrInvalidContent) {
					t.Fatal(err)
				}
			} else if !errors.Is(err, want) {
				t.Fatal(err)
			}
			entries, err := os.ReadDir(filepath.Join(l.filepath, ".metadata"))
			if err != nil || len(entries) != 0 {
				t.Fatalf("metadata upload leaked: %v %v", entries, err)
			}
		})
	}
}

func TestSaveFileWriteAndCloseFailures(t *testing.T) {
	for _, stage := range []string{"write", "close"} {
		t.Run(stage, func(t *testing.T) {
			ops := systemFiles
			want := errors.New(stage)
			ops.OpenFile = func(path string, flags int, mode fs.FileMode) (writableFile, error) {
				f, err := os.OpenFile(path, flags, mode)
				if err != nil {
					return nil, err
				}
				wrapped := &failingFile{File: f}
				if stage == "write" {
					wrapped.writeErr = want
				} else {
					wrapped.closeErr = want
				}
				return wrapped, nil
			}
			if err := saveFileWithFiles(&ops, filepath.Join(t.TempDir(), "file"), []byte("data")); !errors.Is(err, want) {
				t.Fatal(err)
			}
		})
	}
}

type failingEntry struct {
	os.DirEntry
	err error
}

func (e failingEntry) Info() (fs.FileInfo, error) { return nil, e.err }
func TestInventoryEntryErrors(t *testing.T) {
	for _, disappeared := range []bool{false, true} {
		l := mediaFixture(t)
		ops := systemFiles
		l.fileOps = &ops
		path := filepath.Join(l.filepath, "files", strings.Repeat("a", 40))
		writeTestFile(t, path, []byte("data"))
		want := errors.New("entry stat failed")
		if disappeared {
			want = os.ErrNotExist
		}
		ops.ReadDir = func(path string) ([]os.DirEntry, error) {
			entries, err := os.ReadDir(path)
			if len(entries) > 0 {
				entries[0] = failingEntry{DirEntry: entries[0], err: want}
			}
			return entries, err
		}
		inventory, err := l.MediaInventory()
		if disappeared {
			if err != nil || inventory.Count != 0 {
				t.Fatal(inventory, err)
			}
		} else if !errors.Is(err, want) {
			t.Fatal(err)
		}
	}
}

func TestMediaOperationFailures(t *testing.T) {
	for _, stage := range []string{"lookup", "save-stat", "save-bytes", "save-metadata", "inventory", "delete-stat", "delete-remove"} {
		t.Run(stage, func(t *testing.T) {
			l := mediaFixture(t)
			url, err := l.SaveFile("files", []byte("data"))
			if err != nil {
				t.Fatal(err)
			}
			id := strings.TrimPrefix(url, "/files/")
			path := filepath.Join(l.filepath, id)
			ops := systemFiles
			l.fileOps = &ops
			want := errors.New(stage)
			switch stage {
			case "lookup":
				ops.Lstat = func(name string) (fs.FileInfo, error) {
					if name == path {
						return nil, want
					}
					return os.Lstat(name)
				}
				if _, err := l.mediaPath(id); !errors.Is(err, want) {
					t.Fatal(err)
				}
			case "save-stat":
				ops.Lstat = func(name string) (fs.FileInfo, error) {
					if name == path {
						return nil, want
					}
					return os.Lstat(name)
				}
				if _, err := l.SaveFile("files", []byte("data")); !errors.Is(err, want) {
					t.Fatal(err)
				}
			case "save-bytes":
				ops.Stat = func(string) (fs.FileInfo, error) { return nil, want }
				if _, err := l.SaveFile("files", []byte("data")); !errors.Is(err, want) {
					t.Fatal(err)
				}
			case "save-metadata":
				ops.MkdirAll = func(string, fs.FileMode) error { return want }
				if _, err := l.SaveFile("files", []byte("data")); !errors.Is(err, want) {
					t.Fatal(err)
				}
			case "inventory":
				ops.Lstat = func(string) (fs.FileInfo, error) { return nil, want }
				if _, err := l.MediaInventory(); !errors.Is(err, want) {
					t.Fatal(err)
				}
			default:
				if stage == "delete-stat" {
					ops.Stat = func(string) (fs.FileInfo, error) { return nil, want }
				} else {
					ops.Remove = func(string) error { return want }
				}
				result, err := l.DeleteMedia([]string{id}, "")
				if err != nil || result.Deleted != 0 || len(result.Failed) != 1 || result.Failed[0] != id {
					t.Fatal(result, err)
				}
				if _, err := os.Stat(path); err != nil {
					t.Fatal("failed deletion removed data", err)
				}
			}
		})
	}
}

func TestMCPCredentialRandomFailures(t *testing.T) {
	original := randReader
	t.Cleanup(func() { randReader = original })
	want := errors.New("entropy unavailable")
	l := &Logic{db: &credentialDB{user: &model.User{UID: "u", Flags: 1}}}
	for _, failAt := range []int{0, 1} {
		calls := 0
		randReader = func(data []byte) (int, error) {
			if calls == failAt {
				return 0, want
			}
			calls++
			return len(data), nil
		}
		if _, raw, err := l.CreateMCPToken("u", "Agent", time.Now().Add(time.Hour).Unix()); !errors.Is(err, want) || raw != "" {
			t.Fatal(raw, err)
		}
	}
}

func TestLegacyInventoryAndUnsafeMediaPaths(t *testing.T) {
	l := &Logic{filepath: t.TempDir()}
	if inventory, err := l.MediaInventory(); err != nil || inventory.Count != 0 {
		t.Fatal(inventory, err)
	}
	l = mediaFixture(t)
	for _, dir := range []string{"images", "audios"} {
		writeTestFile(t, filepath.Join(l.filepath, dir, strings.Repeat("a", 40)), []byte("data"))
		writeTestFile(t, filepath.Join(l.filepath, dir, "not-an-attachment"), nil)
	}
	inventory, err := l.MediaInventory()
	if err != nil || inventory.Count != 2 {
		t.Fatal(inventory, err)
	}
	for _, m := range inventory.Items {
		if !m.Legacy || m.Kind == "file" {
			t.Fatal(m)
		}
	}
	dir := filepath.Join(l.filepath, "files")
	if err := os.Remove(dir); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, dir, nil)
	if _, err := l.mediaPath("files/" + strings.Repeat("a", 40)); !errors.Is(err, ErrInvalidContent) {
		t.Fatal(err)
	}
	l = mediaFixture(t)
	payload := []byte("data")
	hash := sha1.Sum(payload)
	path := filepath.Join(l.filepath, "files", hex.EncodeToString(hash[:]))
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := l.SaveFile("files", payload); !errors.Is(err, ErrInvalidContent) {
		t.Fatal(err)
	}
	if mediaName("a\x00\x7fb") != "ab" {
		t.Fatal("control bytes retained")
	}
	id := "audios/" + strings.Repeat("b", 40)
	if err := l.saveMediaMetadata(id, []byte{0, 255, 0, 255}, nil); err != nil {
		t.Fatal(err)
	}
	m, err := l.mediaMetadata(id)
	if err != nil || m.Kind != "audio" {
		t.Fatal(m, err)
	}
}

type failedOptionDB struct {
	model.DB
	closed bool
}

func (d *failedOptionDB) GetOption(string, interface{}) error { return os.ErrPermission }
func (d *failedOptionDB) Close()                              { d.closed = true }
func TestNodeSecretReadFailureClosesDatabase(t *testing.T) {
	db := &failedOptionDB{}
	l := &Logic{db: db}
	if err := l.loadNodeSecret(); !errors.Is(err, os.ErrPermission) || !db.closed || l.db != nil {
		t.Fatal(err, db.closed)
	}
}

type failedWatcher struct{}

func (failedWatcher) Add(string) error { return nil }
func (failedWatcher) Close() error     { return os.ErrClosed }
func TestPluginWatcherCloseFailure(t *testing.T) {
	p := &pluginManager{watcher: failedWatcher{}}
	p.Close()
	if p.watcher != nil {
		t.Fatal("failed watcher retained")
	}
}
