package usecase

import (
	"context"

	"github.com/inraise/asperitas/internal/domain"
)

type UserRepository interface {
	GetUserPosts(_ context.Context, username string) domain.PostList
}

type UserUseCase struct {
	repo UserRepository
}

func NewUserUseCase(repo UserRepository) *UserUseCase {
	return &UserUseCase{
		repo: repo,
	}
}

func (u *UserUseCase) GetUserPosts(ctx context.Context, username string) domain.PostList {
	return u.repo.GetUserPosts(ctx, username)
}
