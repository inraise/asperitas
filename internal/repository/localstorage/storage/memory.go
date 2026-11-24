package memorystorage

import (
	"sync"

	"github.com/inraise/asperitas/internal/domain"
)

type UserRecord struct {
	User         domain.User
	PasswordHash []byte
}

type Storage struct {
	Mu    sync.RWMutex
	Posts domain.PostList
	Users map[string]UserRecord
}

func NewStorage() *Storage {
	return &Storage{
		Users: make(map[string]UserRecord),
		Posts: make(domain.PostList, 0),
	}
}
