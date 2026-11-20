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

func (h *PostsHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	postsList, err := h.uc.GetPosts(h.ctx)
	if err != nil {
		writeJSON(w, 500, err)
		return
	}

	writeJSON(w, 200, postsList)
}

func (h *PostsHandler) GetByCategory(w http.ResponseWriter, r *http.Request) {
	category := r.PathValue("category")
	postsList, err := h.uc.GetByCategory(h.ctx, category)
	if err != nil {
		writeJSON(w, 500, err)
		return
	}

	writeJSON(w, 200, postsList)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
