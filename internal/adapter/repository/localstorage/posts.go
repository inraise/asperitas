package localstorage

import (
	"context"

	"github.com/inraise/asperitas/internal/domain"
	"github.com/inraise/asperitas/internal/storage"
)

type MemoryPostsRepo struct {
	db *storage.Storage
}

func NewMemoryPostsRepo(db *storage.Storage) *MemoryPostsRepo {
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
