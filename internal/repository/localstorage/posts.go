package localstorage

import (
	"context"
	"slices"

	"github.com/inraise/asperitas/internal/domain"
	memorystorage "github.com/inraise/asperitas/internal/repository/localstorage/storage"
)

type MemoryPostRepo struct {
	db *memorystorage.Storage
}

func NewMemoryPostRepo(db *memorystorage.Storage) *MemoryPostRepo {
	return &MemoryPostRepo{db: db}
}

func (r *MemoryPostRepo) List(_ context.Context) (domain.PostList, error) {
	return r.filter(func(domain.Post) bool { return true }), nil
}

func (r *MemoryPostRepo) ListByCategory(_ context.Context, category string) (domain.PostList, error) {
	return r.filter(func(p domain.Post) bool { return p.Category == category }), nil
}

func (r *MemoryPostRepo) ListByAuthor(_ context.Context, authorID string) (domain.PostList, error) {
	return r.filter(func(p domain.Post) bool { return p.Author.Id == authorID }), nil
}

func (r *MemoryPostRepo) filter(keep func(domain.Post) bool) domain.PostList {
	r.db.Mu.RLock()
	defer r.db.Mu.RUnlock()

	out := make(domain.PostList, 0, len(r.db.Posts))
	for i := len(r.db.Posts) - 1; i >= 0; i-- {
		if keep(r.db.Posts[i]) {
			out = append(out, r.db.Posts[i].Clone())
		}
	}
	return out
}

func (r *MemoryPostRepo) Create(_ context.Context, post domain.Post) error {
	r.db.Mu.Lock()
	defer r.db.Mu.Unlock()

	r.db.Posts = append(r.db.Posts, post.Clone())
	return nil
}

func (r *MemoryPostRepo) Update(_ context.Context, id string, fn func(*domain.Post) error) (domain.Post, error) {
	r.db.Mu.Lock()
	defer r.db.Mu.Unlock()

	i := r.indexLocked(id)
	if i < 0 {
		return domain.Post{}, domain.ErrPostNotFound
	}

	work := r.db.Posts[i].Clone()
	if err := fn(&work); err != nil {
		return domain.Post{}, err
	}
	r.db.Posts[i] = work
	return work.Clone(), nil
}

func (r *MemoryPostRepo) Delete(_ context.Context, id string, check func(domain.Post) error) error {
	r.db.Mu.Lock()
	defer r.db.Mu.Unlock()

	i := r.indexLocked(id)
	if i < 0 {
		return domain.ErrPostNotFound
	}
	if check != nil {
		if err := check(r.db.Posts[i]); err != nil {
			return err
		}
	}
	r.db.Posts = slices.Delete(r.db.Posts, i, i+1)
	return nil
}

func (r *MemoryPostRepo) indexLocked(id string) int {
	for i := range r.db.Posts {
		if r.db.Posts[i].Id == id {
			return i
		}
	}
	return -1
}
