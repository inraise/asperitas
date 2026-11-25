package usecase

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/inraise/asperitas/internal/domain"
)

type PostRepository interface {
	List(ctx context.Context) (domain.PostList, error)
	ListByCategory(ctx context.Context, category string) (domain.PostList, error)
	ListByAuthor(ctx context.Context, authorID string) (domain.PostList, error)
	Create(ctx context.Context, post domain.Post) error
	Update(ctx context.Context, id string, fn func(*domain.Post) error) (domain.Post, error)
	Delete(ctx context.Context, id string, check func(domain.Post) error) error
}

type UserFinder interface {
	FindByUsername(ctx context.Context, username string) (domain.User, error)
}

type PostUseCase struct {
	posts PostRepository
	users UserFinder
}

func NewPostUseCase(posts PostRepository, users UserFinder) *PostUseCase {
	return &PostUseCase{posts: posts, users: users}
}

func (u *PostUseCase) List(ctx context.Context) (domain.PostList, error) {
	posts, err := u.posts.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list posts: %w", err)
	}
	sortByScore(posts)
	return posts, nil
}

func (u *PostUseCase) ListByCategory(ctx context.Context, category string) (domain.PostList, error) {
	posts, err := u.posts.ListByCategory(ctx, category)
	if err != nil {
		return nil, fmt.Errorf("list posts by category: %w", err)
	}
	sortByScore(posts)
	return posts, nil
}

func (u *PostUseCase) ListByUser(ctx context.Context, username string) (domain.PostList, error) {
	user, err := u.users.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	posts, err := u.posts.ListByAuthor(ctx, user.Id)
	if err != nil {
		return nil, fmt.Errorf("list posts by author: %w", err)
	}
	return posts, nil
}

func (u *PostUseCase) Get(ctx context.Context, id string) (domain.Post, error) {
	return u.posts.Update(ctx, id, func(p *domain.Post) error {
		p.Views++
		return nil
	})
}

func (u *PostUseCase) Create(ctx context.Context, author domain.User, req domain.PostRequest) (domain.Post, error) {
	id, err := domain.NewID()
	if err != nil {
		return domain.Post{}, err
	}

	post := domain.Post{
		Id:       id,
		Title:    req.Title,
		Category: req.Category,
		Type:     req.Type,
		Author:   author,
		Created:  now(),
		Votes:    []domain.Vote{},
		Comments: []domain.Comment{},
	}
	switch req.Type {
	case domain.PostTypeText:
		post.Text = req.Text
	case domain.PostTypeLink:
		post.Url = req.Url
	}
	post.SetVote(author.Id, 1)

	if err := u.posts.Create(ctx, post); err != nil {
		return domain.Post{}, fmt.Errorf("create post: %w", err)
	}
	return post.Clone(), nil
}

func (u *PostUseCase) Delete(ctx context.Context, user domain.User, postID string) error {
	return u.posts.Delete(ctx, postID, func(p domain.Post) error {
		if !p.CanBeDeletedBy(user) {
			return domain.ErrForbidden
		}
		return nil
	})
}

func (u *PostUseCase) AddComment(ctx context.Context, author domain.User, postID, body string) (domain.Post, error) {
	id, err := domain.NewID()
	if err != nil {
		return domain.Post{}, err
	}
	comment := domain.Comment{Id: id, Author: author, Body: body, Created: now()}

	return u.posts.Update(ctx, postID, func(p *domain.Post) error {
		p.AddComment(comment)
		return nil
	})
}

func (u *PostUseCase) DeleteComment(ctx context.Context, user domain.User, postID, commentID string) (domain.Post, error) {
	return u.posts.Update(ctx, postID, func(p *domain.Post) error {
		c, ok := p.CommentByID(commentID)
		if !ok {
			return domain.ErrCommentNotFound
		}
		if !p.CanDeleteComment(c, user) {
			return domain.ErrForbidden
		}
		p.RemoveComment(commentID)
		return nil
	})
}

func (u *PostUseCase) Vote(ctx context.Context, user domain.User, postID string, value int) (domain.Post, error) {
	if value < -1 || value > 1 {
		return domain.Post{}, fmt.Errorf("invalid vote value %d", value)
	}
	return u.posts.Update(ctx, postID, func(p *domain.Post) error {
		if value == 0 {
			p.RemoveVote(user.Id)
		} else {
			p.SetVote(user.Id, value)
		}
		return nil
	})
}

func sortByScore(posts domain.PostList) {
	sort.SliceStable(posts, func(i, j int) bool { return posts[i].Score > posts[j].Score })
}

func now() string { return time.Now().UTC().Format(time.RFC3339) }
