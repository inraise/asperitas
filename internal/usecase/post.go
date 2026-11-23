package usecase

import (
	"context"
	"fmt"

	"github.com/inraise/asperitas/internal/core/domain"
)

type PostRepository interface {
	GetPost(ctx context.Context, postId string) (*domain.Post, error)
}

type PostUseCase struct {
	repo PostRepository
}

func NewPostUseCase(repo PostRepository) *PostUseCase {
	return &PostUseCase{
		repo: repo,
	}
}

func (u *PostUseCase) GetPost(ctx context.Context, postId string) (*domain.Post, error) {
	post, err := u.repo.GetPost(ctx, postId)
	if err != nil {
		return nil, fmt.Errorf("usecase error: %w", err)
	}

	return post, nil
}
