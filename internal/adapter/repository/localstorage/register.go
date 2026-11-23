package localstorage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"

	"github.com/inraise/asperitas/internal/core/domain"
	memorystorage "github.com/inraise/asperitas/internal/core/storage"
)

type MemoryUsersRepo struct {
	db *memorystorage.Storage
}

func NewMemoryUsersRepo(db *memorystorage.Storage) *MemoryUsersRepo {
	return &MemoryUsersRepo{db: db}
}

func (r *MemoryUsersRepo) Create(_ context.Context, username string, hash []byte) (domain.User, error) {
	key := strings.ToLower(username)

	r.db.Mu.Lock()
	defer r.db.Mu.Unlock()

	if _, exists := r.db.Users[key]; exists {
		return domain.User{}, domain.ErrUsernameTaken
	}

	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return domain.User{}, err
	}

	user := domain.User{Id: hex.EncodeToString(b), Username: username}
	r.db.Users[key] = memorystorage.UserRecord{User: user, PasswordHash: hash}
	return user, nil
}

func (r *MemoryUsersRepo) GetByUsername(_ context.Context, username string) (domain.User, []byte, error) {
	r.db.Mu.RLock()
	defer r.db.Mu.RUnlock()

	rec, ok := r.db.Users[strings.ToLower(username)]
	if !ok {
		return domain.User{}, nil, domain.ErrUserNotFound
	}
	return rec.User, rec.PasswordHash, nil
}
