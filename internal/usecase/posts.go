package usecase

import (
	"context"
	"fmt"
	"sort"

	"github.com/inraise/asperitas/internal/domain"
)

type PostsRepository interface {
	GetPosts(ctx context.Context) (domain.PostList, error)
}

type PostsUseCase struct {
	repo PostsRepository
}

func NewPostsUseCase(repo PostsRepository) *PostsUseCase {
	return &PostsUseCase{repo: repo}
}

func (u *PostsUseCase) GetPosts(ctx context.Context) (domain.PostList, error) {
	postsList, err := u.repo.GetPosts(ctx)
	if err != nil {
		return nil, fmt.Errorf("usecase error: %w", err)
	}

	sort.Slice(postsList, func(i, j int) bool {
		return postsList[i].Score > postsList[j].Score
	})
	return postsList, err
}
