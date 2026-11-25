package v1

import (
	"net/http"

	"github.com/inraise/asperitas/internal/domain"
)

func (h *PostHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(w, r)
	if !ok {
		return
	}

	var req domain.CommentRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if errs := validateStruct(h.validate, req); len(errs) > 0 {
		writeValidation(w, errs)
		return
	}

	post, err := h.uc.AddComment(r.Context(), user, r.PathValue("postId"), req.Comment)
	if err != nil {
		fail(w, "add comment", err)
		return
	}
	writeJSON(w, http.StatusCreated, post)
}

func (h *PostHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(w, r)
	if !ok {
		return
	}

	post, err := h.uc.DeleteComment(r.Context(), user, r.PathValue("postId"), r.PathValue("commentId"))
	if err != nil {
		fail(w, "delete comment", err)
		return
	}
	writeJSON(w, http.StatusOK, post)
}
