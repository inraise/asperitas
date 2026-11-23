package middleware

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/inraise/asperitas/internal/domain"
)

type ctxKey struct{}

func AuthMiddleware(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			parts := strings.Fields(r.Header.Get("Authorization"))
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				unauthorized(w)
				return
			}

			claims := &domain.JwtClaims{}
			token, err := jwt.ParseWithClaims(parts[1], claims,
				func(*jwt.Token) (any, error) { return secret, nil },
				jwt.WithValidMethods([]string{"HS256"}),
				jwt.WithExpirationRequired(),
				jwt.WithIssuedAt(),
			)
			if err != nil || !token.Valid || claims.User.Id == "" {
				unauthorized(w)
				return
			}

			ctx := context.WithValue(r.Context(), ctxKey{}, claims.User)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserFromContext(ctx context.Context) (domain.User, bool) {
	u, ok := ctx.Value(ctxKey{}).(domain.User)
	return u, ok
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "unauthorized"})
}

func LogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s (%s)", r.Method, r.URL.Path, time.Since(start))
	})
}
