package adapterhttp

import (
	"net/http"
	"strings"
)

type WebHandler struct{}

func NewWebHandler() *WebHandler {
	return &WebHandler{}
}

func (h *WebHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/static/") || r.URL.Path == "/manifest.json" {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, "./static/html/index.html")
}
