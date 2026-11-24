package adapterhttp

import (
	"context"
	"net/http"

	"github.com/inraise/asperitas/internal/usecase"
)

type UserHandler struct {
	uc  *usecase.UserUseCase
	ctx context.Context
}

func NewUserHandler(ctx context.Context, uc *usecase.UserUseCase) *UserHandler {
	return &UserHandler{
		uc:  uc,
		ctx: ctx,
	}
}

func (h *UserHandler) GetUserPosts(w http.ResponseWriter, r *http.Request) {
	query := r.PathValue("username")
	posts := h.uc.GetUserPosts(h.ctx, query)

	writeJSON(w, 200, posts)
}
