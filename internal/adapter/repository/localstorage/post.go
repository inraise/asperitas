package localstorage

import (
	"context"
	"fmt"

	"github.com/inraise/asperitas/internal/domain"
	"github.com/inraise/asperitas/internal/storage"
)

type MemoryPostRepo struct {
	db *storage.Storage
}

func NewMemoryPostRepo(db *storage.Storage) *MemoryPostRepo {
	return &MemoryPostRepo{
		db: db,
	}
}

func (r *MemoryPostRepo) GetPost(_ context.Context, postId string) (*domain.Post, error) {
	r.db.Mu.RLock()
	defer r.db.Mu.RUnlock()

	for i, p := range r.db.Posts {
		if p.Id == postId {
			r.db.Posts[i].Views += 1
			return &p, nil
		}
	}

	return nil, fmt.Errorf("don't find post")
}
