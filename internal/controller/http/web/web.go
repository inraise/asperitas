package web

import (
	"net/http"
	"path/filepath"
	"strings"
)

type Handler struct {
	htmlDir string
}

func NewHandler(staticDir string) *Handler {
	return &Handler{htmlDir: filepath.Join(staticDir, "html")}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path

	if p == "/api" || strings.HasPrefix(p, "/api/") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}` + "\n"))
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	switch p {
	case "/favicon.ico", "/manifest.json":
		http.ServeFile(w, r, filepath.Join(h.htmlDir, filepath.Base(p)))
	default:
		http.ServeFile(w, r, filepath.Join(h.htmlDir, "index.html"))
	}
}
