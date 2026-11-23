package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/inraise/asperitas/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type UsersRepository interface {
	Create(ctx context.Context, username string, passwordHash []byte) (domain.User, error)
	GetByUsername(ctx context.Context, username string) (domain.User, []byte, error)
}

type AuthUseCase struct {
	repo   UsersRepository
	secret []byte
}

func NewAuthUseCase(repo UsersRepository, secret []byte) *AuthUseCase {
	return &AuthUseCase{repo: repo, secret: secret}
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
	cl := domain.JwtClaims{
		User: user,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(7 * 24 * time.Hour)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, cl).SignedString(u.secret)
}

func (u *AuthUseCase) Login(ctx context.Context, c domain.Credentials) (string, error) {
	user, hash, err := u.repo.GetByUsername(ctx, c.Username)
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword(hash, []byte(c.Password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) ||
			errors.Is(err, bcrypt.ErrPasswordTooLong) {
			return "", domain.ErrInvalidPassword
		}
		return "", fmt.Errorf("compare password: %w", err)
	}

	return u.issueToken(user)
}
