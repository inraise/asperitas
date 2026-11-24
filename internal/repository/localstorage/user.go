package localstorage

import (
	"context"

	"github.com/inraise/asperitas/internal/domain"
	memorystorage "github.com/inraise/asperitas/internal/repository/localstorage/storage"
)

type MemoryUserRepo struct {
	db *memorystorage.Storage
}

func NewMemoryUserRepo(db *memorystorage.Storage) *MemoryUserRepo {
	return &MemoryUserRepo{
		db: db,
	}
}

func (r *MemoryUserRepo) GetUserPosts(_ context.Context, username string) domain.PostList {
	r.db.Mu.RLock()
	defer r.db.Mu.RUnlock()

	out := make(domain.PostList, 0)
	for _, p := range r.db.Posts {
		if p.Author.Username == username {
			out = append(out, p)
		}
	}

	return out
}
