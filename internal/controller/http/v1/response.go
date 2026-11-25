package v1

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/inraise/asperitas/internal/controller/http/middleware"
	"github.com/inraise/asperitas/internal/domain"
)

const maxBodyBytes = 1 << 20

type fieldError struct {
	Location string `json:"location"`
	Param    string `json:"param"`
	Msg      string `json:"msg"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	body, err := json.Marshal(v)
	if err != nil {
		log.Printf("encode response: %v", err)
		status, body = http.StatusInternalServerError, []byte(`{"message":"internal error"}`)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(append(body, '\n'))
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"message": msg})
}

func writeValidation(w http.ResponseWriter, errs []fieldError) {
	writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"errors": errs})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "bad request")
		return false
	}
	return true
}

func currentUser(w http.ResponseWriter, r *http.Request) (domain.User, bool) {
	u, ok := middleware.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
	}
	return u, ok
}

func fail(w http.ResponseWriter, op string, err error) {
	switch {
	case errors.Is(err, domain.ErrPostNotFound):
		writeError(w, http.StatusNotFound, "post not found")
	case errors.Is(err, domain.ErrCommentNotFound):
		writeError(w, http.StatusNotFound, "comment not found")
	case errors.Is(err, domain.ErrUserNotFound):
		writeError(w, http.StatusNotFound, "user not found")
	case errors.Is(err, domain.ErrForbidden):
		writeError(w, http.StatusUnauthorized, "unauthorized")
	default:
		log.Printf("%s: %v", op, err)
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}
