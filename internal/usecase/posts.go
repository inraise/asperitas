package usecase

import (
	"context"
	"fmt"
	"sort"

	"github.com/inraise/asperitas/internal/domain"
)

type TasksRepository interface {
	GetPosts(ctx context.Context) (domain.PostList, error)
}

type TaskUseCase struct {
	repo TasksRepository
}

func NewTaskUseCase(repo TasksRepository) *TaskUseCase {
	return &TaskUseCase{repo: repo}
}

func (u *TaskUseCase) GetPosts(ctx context.Context) (domain.PostList, error) {
	postsList, err := u.repo.GetPosts(ctx)
	if err != nil {
		return nil, fmt.Errorf("usecase error: %w", err)
	}

	sort.Slice(postsList, func(i, j int) bool {
		return postsList[i].Score > postsList[j].Score
	})
	return postsList, err
}
