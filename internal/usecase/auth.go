package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"github.com/inraise/asperitas/internal/domain"
)

const tokenTTL = 7 * 24 * time.Hour

type AuthUserRepository interface {
	Create(ctx context.Context, user domain.User, passwordHash []byte) error
	GetCredentials(ctx context.Context, username string) (domain.User, []byte, error)
}

type AuthUseCase struct {
	users  AuthUserRepository
	secret []byte
}

func NewAuthUseCase(users AuthUserRepository, secret []byte) *AuthUseCase {
	return &AuthUseCase{users: users, secret: secret}
}

func (u *AuthUseCase) Register(ctx context.Context, c domain.Credentials) (string, error) {
	id, err := domain.NewID()
	if err != nil {
		return "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(c.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	user := domain.User{Id: id, Username: c.Username}
	if err := u.users.Create(ctx, user, hash); err != nil {
		return "", err
	}
	return u.issueToken(user)
}

func (u *AuthUseCase) Login(ctx context.Context, c domain.Credentials) (string, error) {
	user, hash, err := u.users.GetCredentials(ctx, c.Username)
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword(hash, []byte(c.Password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) || errors.Is(err, bcrypt.ErrPasswordTooLong) {
			return "", domain.ErrInvalidPassword
		}
		return "", fmt.Errorf("compare password: %w", err)
	}
	return u.issueToken(user)
}

func (u *AuthUseCase) issueToken(user domain.User) (string, error) {
	now := time.Now()
	claims := domain.JwtClaims{
		User: user,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenTTL)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(u.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return token, nil
}
