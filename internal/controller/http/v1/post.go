package adapterhttp

import (
	"context"
	"net/http"

	"github.com/inraise/asperitas/internal/usecase"
)

type PostHandler struct {
	uc  *usecase.PostUseCase
	ctx context.Context
}

func NewPostHandler(ctx context.Context, uc *usecase.PostUseCase) *PostHandler {
	return &PostHandler{
		uc:  uc,
		ctx: ctx,
	}
}

func (h *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	query := r.PathValue("postId")
	post, err := h.uc.GetPost(h.ctx, query)
	if err != nil {
		writeJSON(w, 404, err)
		return
	}

	if post.Type == "link" {
		http.Redirect(w, r, *post.Url, 200)
		return
	}

	writeJSON(w, 200, post)
}
