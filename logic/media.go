package logic

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// MediaFile describes one physical attachment; names retain aliases from repeated uploads.
type MediaFile struct {
	ID           string   `json:"id"`
	Names        []string `json:"names"`
	Kind         string   `json:"kind"`
	MIME         string   `json:"mime"`
	Size         int64    `json:"size"`
	Modified     int64    `json:"modified"`
	Uploaded     int64    `json:"uploaded"`
	LastUploaded int64    `json:"last_uploaded"`
	Legacy       bool     `json:"legacy"`
	Revision     int64    `json:"revision"`
}
type MediaInventory struct {
	Items    []MediaFile `json:"items"`
	Count    int         `json:"count"`
	Bytes    int64       `json:"bytes"`
	Snapshot string      `json:"snapshot"`
}
type MediaDeleteResult struct {
	Deleted int      `json:"deleted"`
	Bytes   int64    `json:"bytes"`
	Failed  []string `json:"failed"`
}

var ErrMediaChanged = errors.New("media inventory changed")

func mediaID(id string) (string, string, bool) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 || (parts[0] != "images" && parts[0] != "audios" && parts[0] != "files") || len(parts[1]) != 40 {
		return "", "", false
	}
	raw, err := hex.DecodeString(parts[1])
	if err != nil || hex.EncodeToString(raw) != parts[1] {
		return "", "", false
	}
	return parts[0], parts[1], true
}

// Only ordinary directories and files within the configured storage are managed.
func (l *Logic) mediaPath(id string) (string, error) {
	dir, name, ok := mediaID(id)
	if !ok {
		return "", ErrInvalidContent
	}
	info, err := os.Lstat(filepath.Join(l.filepath, dir))
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", ErrInvalidContent
	}
	path := filepath.Join(l.filepath, dir, name)
	info, err = os.Lstat(path)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", ErrInvalidContent
	}
	return path, nil
}
func (l *Logic) metadataPath(id string) string {
	return filepath.Join(l.filepath, ".metadata", strings.ReplaceAll(id, "/", "-")+".json")
}
func (l *Logic) mediaMetadata(id string) (MediaFile, error) {
	var m MediaFile
	dirInfo, dirErr := os.Lstat(filepath.Join(l.filepath, ".metadata"))
	if dirErr != nil || !dirInfo.IsDir() {
		return m, ErrNotFound
	}
	path := l.metadataPath(id)
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() {
		return m, ErrNotFound
	}
	f, err := os.Open(path)
	if err != nil {
		return m, err
	}
	defer func() { _ = f.Close() }()
	err = json.NewDecoder(io.LimitReader(f, 2<<20)).Decode(&m)
	return m, err
}
func mediaName(name string) string {
	name = filepath.Base(strings.ReplaceAll(name, "\\", "/"))
	name = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, name)
	if name == "." || name == "/" {
		return ""
	}
	return string([]rune(name)[:min(len([]rune(name)), 255)])
}
func (l *Logic) saveMediaMetadata(id string, data []byte, names []string) error {
	m, err := l.mediaMetadata(id)
	now := time.Now()
	if err != nil {
		m = MediaFile{Uploaded: now.Unix(), Names: []string{}}
	}
	m.ID = id
	m.LastUploaded = now.Unix()
	m.Revision = now.UnixNano()
	m.MIME = http.DetectContentType(data[:min(len(data), 512)])
	m.Kind = "file"
	switch {
	case strings.HasPrefix(m.MIME, "image/"):
		m.Kind = "image"
	case strings.HasPrefix(m.MIME, "audio/"):
		m.Kind = "audio"
	case strings.HasPrefix(m.MIME, "video/"):
		m.Kind = "video"
	case strings.HasPrefix(m.MIME, "text/plain"):
		m.Kind = "text"
	}
	// Unknown audio formats remain downloadable without pretending they are playable.
	if strings.HasPrefix(id, "audios/") && m.Kind == "file" {
		m.Kind = "audio"
	}
	for _, name := range names {
		name = mediaName(name)
		if name == "" {
			continue
		}
		found := false
		for _, old := range m.Names {
			if name == old {
				found = true
				break
			}
		}
		if !found {
			m.Names = append(m.Names, name)
		}
	}
	if len(m.Names) > 1024 {
		m.Names = m.Names[len(m.Names)-1024:]
	}
	if len(m.Names) == 0 {
		m.Names = []string{strings.Split(id, "/")[1]}
	}
	dir := filepath.Join(l.filepath, ".metadata")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	info, err := os.Lstat(dir)
	if err != nil || !info.IsDir() {
		return ErrInvalidContent
	}
	encoded, err := json.Marshal(m)
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".upload-")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(f.Name()) }()
	if _, err = f.Write(encoded); err != nil {
		_ = f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), l.metadataPath(id))
}
func (l *Logic) mediaInventory() (MediaInventory, error) {
	result := MediaInventory{Items: []MediaFile{}}
	hash := sha256.New()
	for _, dir := range []string{"images", "audios", "files"} {
		path := filepath.Join(l.filepath, dir)
		info, err := os.Lstat(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return result, err
		}
		if !info.IsDir() {
			return result, ErrInvalidContent
		}
		entries, err := os.ReadDir(path)
		if err != nil {
			return result, err
		}
		for _, entry := range entries {
			id := dir + "/" + entry.Name()
			if _, _, ok := mediaID(id); !ok {
				continue
			}
			info, err := entry.Info()
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return result, err
			}
			if !info.Mode().IsRegular() {
				continue
			}
			m, err := l.mediaMetadata(id)
			if err != nil {
				kind := "file"
				if dir == "images" {
					kind = "image"
				}
				if dir == "audios" {
					kind = "audio"
				}
				m = MediaFile{Kind: kind, Names: []string{}, Legacy: true, MIME: "application/octet-stream"}
			}
			m.ID = id
			m.Size = info.Size()
			m.Modified = info.ModTime().Unix()
			_, _ = fmt.Fprintf(hash, "%s:%d:%d:%d\n", id, info.Size(), info.ModTime().UnixNano(), m.Revision)
			result.Items = append(result.Items, m)
			result.Bytes += m.Size
		}
	}
	result.Count = len(result.Items)
	result.Snapshot = hex.EncodeToString(hash.Sum(nil))
	return result, nil
}
func (l *Logic) MediaInventory() (MediaInventory, error) {
	if !l.CanFileStore() {
		return MediaInventory{}, ErrNoSupportMethod
	}
	l.mediaMu.Lock()
	defer l.mediaMu.Unlock()
	return l.mediaInventory()
}
func (l *Logic) OpenMedia(id string) (*os.File, MediaFile, error) {
	if !l.CanFileStore() {
		return nil, MediaFile{}, ErrNoSupportMethod
	}
	l.mediaMu.Lock()
	defer l.mediaMu.Unlock()
	path, err := l.mediaPath(id)
	if err != nil {
		return nil, MediaFile{}, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, MediaFile{}, err
	}
	m, _ := l.mediaMetadata(id)
	m.ID = id
	return f, m, nil
}

// DeleteMedia validates the entire request before touching files. Clear uses a confirmed snapshot.
func (l *Logic) DeleteMedia(ids []string, snapshot string) (MediaDeleteResult, error) {
	result := MediaDeleteResult{Failed: []string{}}
	if !l.CanFileStore() {
		return result, ErrNoSupportMethod
	}
	l.mediaMu.Lock()
	defer l.mediaMu.Unlock()
	if snapshot != "" {
		inventory, err := l.mediaInventory()
		if err != nil {
			return result, err
		}
		if inventory.Snapshot != snapshot {
			return result, ErrMediaChanged
		}
		ids = []string{}
		for _, m := range inventory.Items {
			ids = append(ids, m.ID)
		}
	}
	unique := map[string]bool{}
	for _, id := range ids {
		if _, _, ok := mediaID(id); !ok {
			return result, ErrInvalidContent
		}
		unique[id] = true
	}
	ordered := make([]string, 0, len(unique))
	for id := range unique {
		ordered = append(ordered, id)
	}
	sort.Strings(ordered)
	for _, id := range ordered {
		path, err := l.mediaPath(id)
		if err != nil {
			result.Failed = append(result.Failed, id)
			continue
		}
		info, err := os.Stat(path)
		if err != nil {
			result.Failed = append(result.Failed, id)
			continue
		}
		if err = os.Remove(path); err != nil {
			result.Failed = append(result.Failed, id)
			continue
		}
		result.Deleted++
		result.Bytes += info.Size()
		// Stale metadata never makes a deleted attachment visible in the inventory.
		_ = os.Remove(l.metadataPath(id))
	}
	return result, nil
}

// Publish complete files atomically so a failed write cannot leave a truncated attachment.
func saveMediaBytes(path string, data []byte) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".upload-")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(f.Name()) }()
	if _, err = f.Write(data); err != nil {
		_ = f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}
