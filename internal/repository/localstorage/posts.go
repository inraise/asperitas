package localstorage

import (
	"context"

	"github.com/inraise/asperitas/internal/domain"
	memorystorage "github.com/inraise/asperitas/internal/repository/localstorage/storage"
)

type MemoryPostsRepo struct {
	db *memorystorage.Storage
}

func NewMemoryPostsRepo(db *memorystorage.Storage) *MemoryPostsRepo {
	return &MemoryPostsRepo{
		db: db,
	}
}

func (r *MemoryPostsRepo) GetPosts(_ context.Context) (domain.PostList, error) {
	r.db.Mu.RLock()
	defer r.db.Mu.RUnlock()

	out := make(domain.PostList, len(r.db.Posts))
	copy(out, r.db.Posts)
	return out, nil
}

func (r *MemoryPostsRepo) GetByCategory(ctx context.Context, category string) (domain.PostList, error) {
	r.db.Mu.RLock()
	defer r.db.Mu.RUnlock()

	out := make(domain.PostList, 0, len(r.db.Posts))
	for _, p := range r.db.Posts {
		if p.Category == category {
			out = append(out, p)
		}
	}

	return out, nil
}

func (r *MemoryPostsRepo) CreatePost(_ context.Context, post domain.Post) error {
	r.db.Mu.Lock()
	defer r.db.Mu.Unlock()

	r.db.Posts = append(r.db.Posts, post)
	return nil
}
