package localstorage

import (
	"context"
	"fmt"

	"github.com/inraise/asperitas/internal/core/domain"
	memorystorage "github.com/inraise/asperitas/internal/core/storage"
)

type MemoryPostRepo struct {
	db *memorystorage.Storage
}

func NewMemoryPostRepo(db *memorystorage.Storage) *MemoryPostRepo {
	return &MemoryPostRepo{
		db: db,
	}
}

func (r *MemoryPostRepo) GetPost(_ context.Context, postId string) (*domain.Post, error) {
	r.db.Mu.Lock()
	defer r.db.Mu.Unlock()

	for i := range r.db.Posts {
		if r.db.Posts[i].Id == postId {
			r.db.Posts[i].Views++
			p := r.db.Posts[i]
			return &p, nil
		}
	}
	return nil, fmt.Errorf("don't find post")
}
