package adapterhttp

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/go-playground/validator/v10"
	"github.com/inraise/asperitas/internal/core/domain"
	"github.com/inraise/asperitas/internal/core/middleware"
	"github.com/inraise/asperitas/internal/usecase"
)

type PostsHandler struct {
	uc       *usecase.PostsUseCase
	ctx      context.Context
	validate *validator.Validate
}

func NewPostsHandler(ctx context.Context, uc *usecase.PostsUseCase) *PostsHandler {
	return &PostsHandler{uc: uc, ctx: ctx, validate: newValidator()}
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

func (h *PostsHandler) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "unauthorized"})
		return
	}

	var req domain.PostRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "bad request"})
		return
	}

	if errs := h.validatePost(req); len(errs) > 0 {
		writeValidation(w, errs...)
		return
	}

	post, err := h.uc.CreatePost(r.Context(), user, req)
	if err != nil {
		log.Printf("create post: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "internal error"})
		return
	}
	writeJSON(w, http.StatusCreated, post)
}

func (h *PostsHandler) validatePost(req domain.PostRequest) []fieldError {
	var out []fieldError

	if err := h.validate.Struct(req); err != nil {
		var verrs validator.ValidationErrors
		if errors.As(err, &verrs) {
			for _, fe := range verrs {
				out = append(out, fieldError{"body", fe.Field(), message(fe)})
			}
		}
	}

	if strings.TrimSpace(req.Title) != req.Title {
		out = append(out, fieldError{"body", "title", "must not start or end with whitespace"})
	}

	switch req.Type {
	case "text":
		if req.Text == nil {
			out = append(out, fieldError{"body", "text", "is required"})
		} else if utf8.RuneCountInString(*req.Text) < 4 {
			out = append(out, fieldError{"body", "text", "must be at least 4 characters long"})
		}
	case "link":
		if req.Url == nil {
			out = append(out, fieldError{"body", "url", "is required"})
		} else if u, err := url.ParseRequestURI(*req.Url); err != nil || u.Scheme == "" || u.Host == "" {
			out = append(out, fieldError{"body", "url", "must be a valid absolute URL"})
		}
	}
	return out
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
