package app

import (
	"net/http"

	"github.com/inraise/asperitas/internal/controller/http/middleware"
	v1 "github.com/inraise/asperitas/internal/controller/http/v1"
	"github.com/inraise/asperitas/internal/controller/http/web"
	"github.com/inraise/asperitas/internal/repository/localstorage"
	memorystorage "github.com/inraise/asperitas/internal/repository/localstorage/storage"
	"github.com/inraise/asperitas/internal/usecase"
)

func New(jwtSecret []byte, staticDir string) http.Handler {
	db := memorystorage.NewStorage()
	userRepo := localstorage.NewMemoryUserRepo(db)
	postRepo := localstorage.NewMemoryPostRepo(db)

	authUC := usecase.NewAuthUseCase(userRepo, jwtSecret)
	postUC := usecase.NewPostUseCase(postRepo, userRepo)

	authH := v1.NewAuthHandler(authUC)
	postH := v1.NewPostHandler(postUC)

	authRequired := middleware.AuthMiddleware(jwtSecret)
	protected := func(h http.HandlerFunc) http.Handler { return authRequired(h) }

	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	mux.Handle("/{$}", web.NewHandler(staticDir))

	// Аутентификация
	mux.HandleFunc("POST /api/register", authH.Register)
	mux.HandleFunc("POST /api/login", authH.Login)

	// Списки
	mux.HandleFunc("GET /api/posts", postH.List)
	mux.HandleFunc("GET /api/posts/{$}", postH.List)
	mux.HandleFunc("GET /api/posts/{category}", postH.ListByCategory)
	mux.HandleFunc("GET /api/user/{username}", postH.ListByUser)

	// Создание поста
	mux.Handle("POST /api/posts", protected(postH.Create))
	mux.Handle("POST /api/posts/{$}", protected(postH.Create))

	// Один пост
	mux.HandleFunc("GET /api/post/{postId}", postH.Get)
	mux.Handle("DELETE /api/post/{postId}", protected(postH.Delete))

	// Комментарии
	mux.Handle("POST /api/post/{postId}", protected(postH.AddComment))
	mux.Handle("DELETE /api/post/{postId}/{commentId}", protected(postH.DeleteComment))

	// Голоса
	mux.Handle("GET /api/post/{postId}/upvote", protected(postH.Upvote))
	mux.Handle("GET /api/post/{postId}/downvote", protected(postH.Downvote))
	mux.Handle("GET /api/post/{postId}/unvote", protected(postH.Unvote))

	return middleware.LogMiddleware(mux)
}
