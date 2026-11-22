package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/inraise/asperitas/internal/domain"
)

type UsersRepository interface {
	Create(ctx context.Context, username string, passwordHash []byte) (domain.User, error)
}

type AuthUseCase struct {
	repo   UsersRepository
	secret []byte
}

func NewAuthUseCase(repo UsersRepository, secret []byte) *AuthUseCase {
	return &AuthUseCase{repo: repo, secret: secret}
}

type claims struct {
	User domain.User `json:"user"`
	jwt.RegisteredClaims
}

func (u *AuthUseCase) Register(ctx context.Context, c domain.Credentials) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(c.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	user, err := u.repo.Create(ctx, c.Username, hash)
	if err != nil {
		return "", err
	}

	return u.issueToken(user)
}

func (u *AuthUseCase) issueToken(user domain.User) (string, error) {
	now := time.Now()
	cl := claims{
		User: user,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(7 * 24 * time.Hour)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, cl).SignedString(u.secret)
}
