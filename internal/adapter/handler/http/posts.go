package adapterhttp

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/inraise/asperitas/internal/usecase"
)

type PostsHandler struct {
	uc  *usecase.PostsUseCase
	ctx context.Context
}

func NewPostsHandler(ctx context.Context, uc *usecase.PostsUseCase) *PostsHandler {
	return &PostsHandler{uc: uc, ctx: ctx}
}

func (h *PostsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		postsList, err := h.uc.GetPosts(h.ctx)
		if err != nil {
			writeJSON(w, 500, err)
			return
		}

		writeJSON(w, 200, postsList)
	default:
		http.Error(w, "method not allowed", 405)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
