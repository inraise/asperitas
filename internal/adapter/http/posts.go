package adapterhttp

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/inraise/asperitas/internal/usecase"
)

type TaskHandler struct {
	uc  *usecase.TaskUseCase
	ctx context.Context
}

func NewTaskHandler(ctx context.Context, uc *usecase.TaskUseCase) *TaskHandler {
	return &TaskHandler{uc: uc, ctx: ctx}
}

func (h *TaskHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		tasksList, err := h.uc.GetPosts(h.ctx)
		if err != nil {
			writeJSON(w, 500, err)
			return
		}

		writeJSON(w, 200, tasksList)
	default:
		http.Error(w, "method not allowed", 405)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
