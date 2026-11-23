package memorystorage

import (
	"sync"
	"time"

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
		Posts: domain.PostList{
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
