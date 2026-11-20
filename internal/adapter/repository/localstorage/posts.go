package localstorage

import (
	"context"
	"sync"
	"time"

	"github.com/inraise/asperitas/internal/domain"
)

type MemoryTaskRepo struct {
	mu    sync.RWMutex
	tasks domain.PostList
}

func NewMemoryTaskRepo() *MemoryTaskRepo {
	return &MemoryTaskRepo{
		tasks: domain.PostList{
			domain.Post{
				Id:               "1",
				Title:            "Первый тестовый пост",
				Category:         "music",
				Type:             "text",
				Text:             ptr("Привет!"),
				Author:           domain.User{Id: "u1", Username: "denis"},
				Score:            1,
				Votes:            []domain.Vote{{User: "u2", Vote: 1}},
				Comments:         []domain.Comment{},
				Created:          time.Now().UTC().Format(time.RFC3339),
				UpvotePercentage: 100,
			},
			domain.Post{
				Id:               "2",
				Title:            "Второй тестовый пост",
				Category:         "news",
				Type:             "link",
				Url:              ptr("google.com"),
				Author:           domain.User{Id: "u2", Username: "alice"},
				Score:            4,
				Votes:            []domain.Vote{{User: "u1", Vote: -1}},
				Comments:         []domain.Comment{},
				Created:          time.Now().UTC().Format(time.RFC3339),
				UpvotePercentage: 100,
			},
		},
	}
}

func ptr(s string) *string {
	return &s
}

func (r *MemoryTaskRepo) GetPosts(_ context.Context) (domain.PostList, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make(domain.PostList, len(r.tasks))
	copy(out, r.tasks)
	return out, nil
}
