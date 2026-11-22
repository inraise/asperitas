package adapterhttp

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/inraise/asperitas/internal/domain"
	"github.com/inraise/asperitas/internal/usecase"
)

type AuthHandler struct {
	uc       *usecase.AuthUseCase
	validate *validator.Validate
}

func NewAuthHandler(uc *usecase.AuthUseCase) *AuthHandler {
	v := validator.New()

	v.RegisterTagNameFunc(func(f reflect.StructField) string {
		return strings.Split(f.Tag.Get("json"), ",")[0]
	})
	return &AuthHandler{uc: uc, validate: v}
}

type fieldError struct {
	Location string `json:"location"`
	Param    string `json:"param"`
	Msg      string `json:"msg"`
}

func writeValidation(w http.ResponseWriter, errs ...fieldError) {
	writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"errors": errs})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var creds domain.Credentials
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	if err := dec.Decode(&creds); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "bad request"})
		return
	}

	if err := h.validate.Struct(creds); err != nil {
		var verrs validator.ValidationErrors
		if !errors.As(err, &verrs) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "bad request"})
			return
		}
		out := make([]fieldError, 0, len(verrs))
		for _, fe := range verrs {
			out = append(out, fieldError{"body", fe.Field(), message(fe)})
		}
		writeValidation(w, out...)
		return
	}

	token, err := h.uc.Register(r.Context(), creds)
	switch {
	case errors.Is(err, domain.ErrUsernameTaken):
		writeValidation(w, fieldError{"body", "username", "already exists"})
	case err != nil:
		log.Printf("register: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "internal error"})
	default:
		writeJSON(w, http.StatusCreated, domain.AuthenticationResponse{Token: token})
	}
}

func message(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "min":
		return "must be at least " + fe.Param() + " characters long"
	case "max":
		return "must be at most " + fe.Param() + " characters long"
	default:
		return "is invalid"
	}
}
