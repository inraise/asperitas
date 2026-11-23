package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"time"

	"github.com/inraise/asperitas/internal/core/domain"
)

type PostsRepository interface {
	GetPosts(ctx context.Context) (domain.PostList, error)
	GetByCategory(ctx context.Context, category string) (domain.PostList, error)
	CreatePost(ctx context.Context, post domain.Post) error
}

type PostsUseCase struct {
	repo PostsRepository
}

func NewPostsUseCase(repo PostsRepository) *PostsUseCase {
	return &PostsUseCase{repo: repo}
}

func (u *PostsUseCase) GetPosts(ctx context.Context) (domain.PostList, error) {
	postsList, err := u.repo.GetPosts(ctx)
	if err != nil {
		return nil, fmt.Errorf("usecase error: %w", err)
	}

	sort.Slice(postsList, func(i, j int) bool {
		return postsList[i].Score > postsList[j].Score
	})
	return postsList, err
}

func (u *PostsUseCase) GetByCategory(ctx context.Context, category string) (domain.PostList, error) {
	postsList, err := u.repo.GetByCategory(ctx, category)
	if err != nil {
		return nil, fmt.Errorf("usecase error: %w", err)
	}

	sort.Slice(postsList, func(i, j int) bool {
		return postsList[i].Score > postsList[j].Score
	})
	return postsList, err
}

func (u *PostsUseCase) CreatePost(ctx context.Context, author domain.User, req domain.PostRequest) (domain.Post, error) {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return domain.Post{}, fmt.Errorf("generate id: %w", err)
	}

	post := domain.Post{
		Id:               hex.EncodeToString(b),
		Title:            req.Title,
		Category:         req.Category,
		Type:             req.Type,
		Author:           author,
		Score:            1,
		Views:            0,
		Votes:            []domain.Vote{{User: author.Id, Vote: 1}},
		Comments:         []domain.Comment{},
		Created:          time.Now().UTC().Format(time.RFC3339),
		UpvotePercentage: 100,
	}
	switch req.Type {
	case "text":
		post.Text = req.Text
	case "link":
		post.Url = req.Url
	}

	if err := u.repo.CreatePost(ctx, post); err != nil {
		return domain.Post{}, fmt.Errorf("usecase error: %w", err)
	}
	return post, nil
}
