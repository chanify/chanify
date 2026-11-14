package core

import (
	"io"
	"net/http"
	"strings"
	"sync"
)

// Serialize SSE comments with SDK writes. JSON responses never receive comments.
type mcpStreamWriter struct {
	http.ResponseWriter
	mu        sync.Mutex
	started   bool
	streaming bool
	onStream  func()
}

func (w *mcpStreamWriter) start() {
	if !w.started {
		w.started = true
		w.streaming = strings.HasPrefix(w.Header().Get("Content-Type"), "text/event-stream")
		if w.streaming && w.onStream != nil {
			w.onStream()
		}
	}
}
func (w *mcpStreamWriter) WriteHeader(code int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.start()
	w.ResponseWriter.WriteHeader(code)
}
func (w *mcpStreamWriter) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.start()
	return w.ResponseWriter.Write(data)
}
func (w *mcpStreamWriter) Flush() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.start()
	_ = http.NewResponseController(w.ResponseWriter).Flush()
}
func (w *mcpStreamWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }
func (w *mcpStreamWriter) heartbeat() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.streaming {
		_, _ = io.WriteString(w.ResponseWriter, ": keepalive\n\n")
		_ = http.NewResponseController(w.ResponseWriter).Flush()
	}
}
