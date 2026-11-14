package core

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/chanify/chanify/logic"
)

func (c *Core) adminMediaList(w http.ResponseWriter, r *http.Request) {
	if !c.logic.CanFileStore() {
		http.NotFound(w, r)
		return
	}
	inventory, err := c.logic.MediaInventory()
	if err != nil {
		adminJSON(w, 500, map[string]string{"error": "无法读取文件列表"})
		return
	}
	q := r.URL.Query()
	search := strings.ToLower(strings.TrimSpace(q.Get("q")))
	kind := q.Get("kind")
	items := []logic.MediaFile{}
	for _, m := range inventory.Items {
		if kind != "" && kind != m.Kind {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(m.ID+" "+strings.Join(m.Names, " ")), search) {
			continue
		}
		items = append(items, m)
	}
	sortBy := q.Get("sort")
	sort.Slice(items, func(i, j int) bool {
		a, b := items[i], items[j]
		switch sortBy {
		case "size":
			if a.Size != b.Size {
				return a.Size > b.Size
			}
		case "oldest":
			ta, tb := mediaTime(a), mediaTime(b)
			if ta != tb {
				return ta < tb
			}
		case "name":
			na, nb := mediaDisplayName(a), mediaDisplayName(b)
			if na != nb {
				return na < nb
			}
		default:
			ta, tb := mediaTime(a), mediaTime(b)
			if ta != tb {
				return ta > tb
			}
		}
		return a.ID < b.ID
	})
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	limit := 50
	total := len(items)
	pages := max(1, (total+limit-1)/limit)
	if page > pages {
		page = pages
	}
	start := (page - 1) * limit
	var filteredBytes int64
	for _, m := range items {
		filteredBytes += m.Size
	}
	adminJSON(w, 200, map[string]interface{}{"items": items[start:min(start+limit, total)], "total": total, "filtered_bytes": filteredBytes, "page": page, "pages": pages, "count": inventory.Count, "bytes": inventory.Bytes, "snapshot": inventory.Snapshot})
}
func mediaDisplayName(m logic.MediaFile) string {
	if len(m.Names) > 0 {
		return strings.ToLower(m.Names[len(m.Names)-1])
	}
	return m.ID
}
func mediaTime(m logic.MediaFile) int64 {
	if m.Uploaded != 0 {
		return m.Uploaded
	}
	return m.Modified
}
func (c *Core) adminMediaDelete(w http.ResponseWriter, r *http.Request) {
	if !c.logic.CanFileStore() {
		http.NotFound(w, r)
		return
	}
	var p struct {
		Action   string   `json:"action"`
		IDs      []string `json:"ids"`
		Snapshot string   `json:"snapshot"`
	}
	if adminDecode(w, r, &p) != nil || (p.Action != "delete" && p.Action != "clear") || (p.Action == "delete" && (len(p.IDs) == 0 || len(p.IDs) > 50 || p.Snapshot != "")) || (p.Action == "clear" && (p.Snapshot == "" || len(p.IDs) != 0)) {
		adminJSON(w, 400, map[string]string{"error": "参数无效"})
		return
	}
	result, err := c.logic.DeleteMedia(p.IDs, p.Snapshot)
	if errors.Is(err, logic.ErrMediaChanged) {
		adminJSON(w, 409, map[string]string{"error": "文件列表已变化，请刷新后重新确认"})
		return
	}
	if err != nil {
		adminJSON(w, 400, map[string]string{"error": "无法删除文件"})
		return
	}
	adminJSON(w, 200, result)
}

// Admin downloads use the existing session, never client sending credentials.
func (c *Core) adminMediaContent(w http.ResponseWriter, r *http.Request) {
	if !c.logic.CanFileStore() {
		http.NotFound(w, r)
		return
	}
	f, m, err := c.logic.OpenMedia(r.URL.Query().Get("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil {
		http.NotFound(w, r)
		return
	}
	name := strings.ReplaceAll(m.ID, "/", "-")
	if len(m.Names) > 0 {
		name = m.Names[len(m.Names)-1]
	}
	contentType := "application/octet-stream"
	disposition := "attachment"
	if r.URL.Query().Get("preview") == "1" {
		data := make([]byte, 512)
		n, _ := f.Read(data)
		_, _ = f.Seek(0, io.SeekStart)
		detected := http.DetectContentType(data[:n])
		// Deliberately exclude HTML, SVG and documents with active content.
		switch detected {
		case "image/png", "image/jpeg", "image/gif", "image/webp", "image/bmp", "image/x-icon", "audio/mpeg", "audio/wave", "audio/wav", "audio/aiff", "audio/ogg", "application/ogg", "video/mp4", "video/webm":
			contentType = detected
			disposition = "inline"
		default:
			if strings.HasPrefix(detected, "text/plain") {
				contentType = "text/plain; charset=utf-8"
				disposition = "inline"
			}
		}
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", mime.FormatMediaType(disposition, map[string]string{"filename": name}))
	w.Header().Set("Content-Security-Policy", "default-src 'none'; sandbox; frame-ancestors 'none'")
	if strings.HasPrefix(contentType, "text/plain") {
		_, _ = io.Copy(w, io.LimitReader(f, 256<<10))
		return
	}
	http.ServeContent(w, r, name, info.ModTime(), f)
}
