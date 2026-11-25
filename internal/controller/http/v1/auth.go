package v1

import (
	"errors"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/inraise/asperitas/internal/domain"
	"github.com/inraise/asperitas/internal/usecase"
)

type AuthHandler struct {
	uc       *usecase.AuthUseCase
	validate *validator.Validate
}

func NewAuthHandler(uc *usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{uc: uc, validate: newValidator()}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	creds, ok := h.readCredentials(w, r)
	if !ok {
		return
	}

	token, err := h.uc.Register(r.Context(), creds)
	switch {
	case errors.Is(err, domain.ErrUsernameTaken):
		writeValidation(w, []fieldError{{Location: "body", Param: "username", Msg: "already exists"}})
	case err != nil:
		log.Printf("register: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
	default:
		writeJSON(w, http.StatusCreated, domain.AuthenticationResponse{Token: token})
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	creds, ok := h.readCredentials(w, r)
	if !ok {
		return
	}

	token, err := h.uc.Login(r.Context(), creds)
	switch {
	case errors.Is(err, domain.ErrUserNotFound), errors.Is(err, domain.ErrInvalidPassword):
		writeError(w, http.StatusUnauthorized, err.Error())
	case err != nil:
		log.Printf("login: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
	default:
		writeJSON(w, http.StatusOK, domain.AuthenticationResponse{Token: token})
	}
}

func (h *AuthHandler) readCredentials(w http.ResponseWriter, r *http.Request) (domain.Credentials, bool) {
	var creds domain.Credentials
	if !decodeJSON(w, r, &creds) {
		return creds, false
	}
	if errs := validateStruct(h.validate, creds); len(errs) > 0 {
		writeValidation(w, errs)
		return creds, false
	}
	return creds, true
}
