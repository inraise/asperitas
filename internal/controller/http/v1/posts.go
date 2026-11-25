package v1

import (
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/go-playground/validator/v10"
	"github.com/inraise/asperitas/internal/domain"
	"github.com/inraise/asperitas/internal/usecase"
)

type PostHandler struct {
	uc       *usecase.PostUseCase
	validate *validator.Validate
}

func NewPostHandler(uc *usecase.PostUseCase) *PostHandler {
	return &PostHandler{uc: uc, validate: newValidator()}
}

func (h *PostHandler) List(w http.ResponseWriter, r *http.Request) {
	posts, err := h.uc.List(r.Context())
	if err != nil {
		fail(w, "list posts", err)
		return
	}
	writeJSON(w, http.StatusOK, posts)
}

func (h *PostHandler) ListByCategory(w http.ResponseWriter, r *http.Request) {
	posts, err := h.uc.ListByCategory(r.Context(), r.PathValue("category"))
	if err != nil {
		fail(w, "list posts by category", err)
		return
	}
	writeJSON(w, http.StatusOK, posts)
}

func (h *PostHandler) ListByUser(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if !domain.ValidUsername(username) {
		writeError(w, http.StatusBadRequest, "bad request")
		return
	}
	posts, err := h.uc.ListByUser(r.Context(), username)
	if err != nil {
		fail(w, "list posts by user", err)
		return
	}
	writeJSON(w, http.StatusOK, posts)
}

func (h *PostHandler) Get(w http.ResponseWriter, r *http.Request) {
	post, err := h.uc.Get(r.Context(), r.PathValue("postId"))
	if err != nil {
		fail(w, "get post", err)
		return
	}
	writeJSON(w, http.StatusOK, post)
}

func (h *PostHandler) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(w, r)
	if !ok {
		return
	}

	var req domain.PostRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if errs := h.validatePost(req); len(errs) > 0 {
		writeValidation(w, errs)
		return
	}

	post, err := h.uc.Create(r.Context(), user, req)
	if err != nil {
		fail(w, "create post", err)
		return
	}
	writeJSON(w, http.StatusCreated, post)
}

func (h *PostHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(w, r)
	if !ok {
		return
	}
	if err := h.uc.Delete(r.Context(), user, r.PathValue("postId")); err != nil {
		fail(w, "delete post", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "success"})
}

func (h *PostHandler) validatePost(req domain.PostRequest) []fieldError {
	errs := validateStruct(h.validate, req)

	if strings.TrimSpace(req.Title) != req.Title {
		errs = append(errs, fieldError{"body", "title", "must not start or end with whitespace"})
	}

	switch req.Type {
	case domain.PostTypeText:
		if req.Text == nil {
			errs = append(errs, fieldError{"body", "text", "is required"})
		} else if utf8.RuneCountInString(*req.Text) < 4 {
			errs = append(errs, fieldError{"body", "text", "must be at least 4 characters long"})
		}
	case domain.PostTypeLink:
		if req.Url == nil {
			errs = append(errs, fieldError{"body", "url", "is required"})
		} else if !validWebURL(*req.Url) {
			errs = append(errs, fieldError{"body", "url", "must be an absolute http(s) URL"})
		}
	}
	return errs
}

func validWebURL(s string) bool {
	u, err := url.ParseRequestURI(s)
	if err != nil || u.Host == "" {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}
