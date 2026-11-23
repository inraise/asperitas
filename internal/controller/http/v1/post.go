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

func (h *PostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		query :=  r.PathValue("postId")
		post, err := h.uc.GetPost(h.ctx, query)
		if err != nil {
			writeJSON(w, 404, err)
			return
		}

		writeJSON(w, 200, post)
	default:
		http.Error(w, "method not allowed", 405)
	}
}
