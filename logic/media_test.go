package logic

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func mediaFixture(t *testing.T) *Logic {
	t.Helper()
	l := &Logic{filepath: t.TempDir()}
	for _, dir := range []string{"images", "audios", "files"} {
		if err := os.Mkdir(filepath.Join(l.filepath, dir), 0700); err != nil {
			t.Fatal(err)
		}
	}
	return l
}
func TestMediaMetadataAliasesAndLegacy(t *testing.T) {
	l := mediaFixture(t)
	data := []byte("synthetic attachment")
	path, err := l.SaveFile("files", data, "../report.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = l.SaveFile("files", data, "报告.txt"); err != nil {
		t.Fatal(err)
	}
	inventory, err := l.MediaInventory()
	if err != nil {
		t.Fatal(err)
	}
	if inventory.Count != 1 || inventory.Bytes != int64(len(data)) {
		t.Fatal(inventory)
	}
	m := inventory.Items[0]
	if m.Legacy || m.Kind != "text" || m.Uploaded == 0 || len(m.Names) != 2 || m.Names[0] != "report.txt" {
		t.Fatal(m)
	}
	// Metadata survives a new Logic instance without changing the existing download format.
	next := &Logic{filepath: l.filepath}
	again, err := next.MediaInventory()
	if err != nil || len(again.Items[0].Names) != 2 {
		t.Fatal(again, err)
	}
	loaded, err := next.LoadFile("files", strings.TrimPrefix(path, "/files/files/"))
	if err != nil || !bytes.Equal(loaded, data) {
		t.Fatal(err)
	}
	legacy := sha1.Sum([]byte("legacy"))
	legacyID := "files/" + hex.EncodeToString(legacy[:])
	if err = os.WriteFile(filepath.Join(l.filepath, legacyID), []byte("legacy"), 0600); err != nil {
		t.Fatal(err)
	}
	inventory, err = l.MediaInventory()
	if err != nil || inventory.Count != 2 {
		t.Fatal(inventory, err)
	}
	found := false
	for _, item := range inventory.Items {
		if item.ID == legacyID {
			found = item.Legacy && item.Modified > 0 && item.Uploaded == 0
		}
	}
	if !found {
		t.Fatal("legacy file not represented honestly")
	}
	result, err := l.DeleteMedia([]string{legacyID}, "")
	if err != nil || result.Deleted != 1 || result.Bytes != 6 {
		t.Fatal(result, err)
	}
}
func TestMediaClearSnapshotAndPartialDeletion(t *testing.T) {
	l := mediaFixture(t)
	_, _ = l.SaveFile("files", []byte("one"), "one.txt")
	before, _ := l.MediaInventory()
	_, _ = l.SaveFile("files", []byte("one"), "new-alias.txt")
	if _, err := l.DeleteMedia(nil, before.Snapshot); !errors.Is(err, ErrMediaChanged) {
		t.Fatal(err)
	}
	before, _ = l.MediaInventory()
	_, _ = l.SaveFile("files", []byte("two"), "two.txt")
	if _, err := l.DeleteMedia(nil, before.Snapshot); !errors.Is(err, ErrMediaChanged) {
		t.Fatal(err)
	}
	current, _ := l.MediaInventory()
	missing := "files/" + strings.Repeat("0", 40)
	result, err := l.DeleteMedia([]string{current.Items[0].ID, current.Items[0].ID, missing}, "")
	if err != nil || result.Deleted != 1 || len(result.Failed) != 1 || result.Failed[0] != missing {
		t.Fatal(result, err)
	}
	current, _ = l.MediaInventory()
	result, err = l.DeleteMedia(nil, current.Snapshot)
	if err != nil || result.Deleted != 1 {
		t.Fatal(result, err)
	}
	current, _ = l.MediaInventory()
	if current.Count != 0 || current.Bytes != 0 {
		t.Fatal(current)
	}
}
func TestMediaRejectsTraversalAndLinks(t *testing.T) {
	l := mediaFixture(t)
	_, _ = l.SaveFile("files", []byte("safe"), "safe.txt")
	current, _ := l.MediaInventory()
	id := current.Items[0].ID
	for _, bad := range []string{"../outside", "files/../outside", "other/" + strings.Repeat("a", 40), "files/" + strings.Repeat("A", 40)} {
		if _, _, err := l.OpenMedia(bad); err == nil {
			t.Fatal("opened", bad)
		}
		if _, err := l.DeleteMedia([]string{id, bad}, ""); err == nil {
			t.Fatal("deleted invalid batch", bad)
		}
	}
	current, _ = l.MediaInventory()
	if current.Count != 1 {
		t.Fatal("invalid batch deleted valid files")
	}
	outside := filepath.Join(t.TempDir(), "outside")
	if err := os.WriteFile(outside, []byte("private"), 0600); err != nil {
		t.Fatal(err)
	}
	linkID := "files/" + strings.Repeat("b", 40)
	if err := os.Symlink(outside, filepath.Join(l.filepath, linkID)); err != nil {
		t.Fatal(err)
	}
	if _, _, err := l.OpenMedia(linkID); err == nil {
		t.Fatal("opened symlink")
	}
	result, err := l.DeleteMedia([]string{linkID}, "")
	if err != nil || result.Deleted != 0 || len(result.Failed) != 1 {
		t.Fatal(result, err)
	}
	current, _ = l.MediaInventory()
	if current.Count != 1 {
		t.Fatal("listed link")
	}
	if _, err := os.Stat(outside); err != nil {
		t.Fatal("deleted external file", err)
	}
	if err := os.RemoveAll(filepath.Join(l.filepath, "audios")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Dir(outside), filepath.Join(l.filepath, "audios")); err != nil {
		t.Fatal(err)
	}
	if _, err := l.SaveFile("audios", []byte("test")); err == nil {
		t.Fatal("wrote into linked directory")
	}
}
func TestMediaConcurrentAliasesAndDisabledStorage(t *testing.T) {
	l := mediaFixture(t)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if _, err := l.SaveFile("files", []byte("shared"), fmt.Sprintf("report-%d.txt", i)); err != nil {
				t.Error(err)
			}
		}(i)
	}
	wg.Wait()
	inventory, err := l.MediaInventory()
	if err != nil || inventory.Count != 1 || len(inventory.Items[0].Names) != 20 {
		t.Fatal(inventory, err)
	}
	disabled := &Logic{}
	if _, err := disabled.MediaInventory(); err != ErrNoSupportMethod {
		t.Fatal(err)
	}
	if _, _, err := disabled.OpenMedia(inventory.Items[0].ID); err != ErrNoSupportMethod {
		t.Fatal(err)
	}
	if _, err := disabled.DeleteMedia([]string{inventory.Items[0].ID}, ""); err != ErrNoSupportMethod {
		t.Fatal(err)
	}
}
