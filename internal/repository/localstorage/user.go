package localstorage

import (
	"context"
	"strings"

	"github.com/inraise/asperitas/internal/domain"
	memorystorage "github.com/inraise/asperitas/internal/repository/localstorage/storage"
)

type MemoryUserRepo struct {
	db *memorystorage.Storage
}

func NewMemoryUserRepo(db *memorystorage.Storage) *MemoryUserRepo {
	return &MemoryUserRepo{db: db}
}

func (r *MemoryUserRepo) Create(_ context.Context, user domain.User, passwordHash []byte) error {
	key := strings.ToLower(user.Username)

	r.db.Mu.Lock()
	defer r.db.Mu.Unlock()

	if _, exists := r.db.Users[key]; exists {
		return domain.ErrUsernameTaken
	}
	r.db.Users[key] = memorystorage.UserRecord{User: user, PasswordHash: passwordHash}
	return nil
}

func (r *MemoryUserRepo) GetCredentials(_ context.Context, username string) (domain.User, []byte, error) {
	r.db.Mu.RLock()
	defer r.db.Mu.RUnlock()

	rec, ok := r.db.Users[strings.ToLower(username)]
	if !ok {
		return domain.User{}, nil, domain.ErrUserNotFound
	}
	return rec.User, rec.PasswordHash, nil
}

func (r *MemoryUserRepo) FindByUsername(_ context.Context, username string) (domain.User, error) {
	r.db.Mu.RLock()
	defer r.db.Mu.RUnlock()

	rec, ok := r.db.Users[strings.ToLower(username)]
	if !ok {
		return domain.User{}, domain.ErrUserNotFound
	}
	return rec.User, nil
}
