package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	adapterhttp "github.com/inraise/asperitas/internal/adapter/handler/http"
	"github.com/inraise/asperitas/internal/adapter/repository/localstorage"
	"github.com/inraise/asperitas/internal/core/middleware"
	memorystorage "github.com/inraise/asperitas/internal/core/storage"
	"github.com/inraise/asperitas/internal/usecase"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	memDB := memorystorage.NewStorage()
	
	mux := http.NewServeMux()

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET is not set")
	}

	webHandler := adapterhttp.NewWebHandler()

	usersRepo := localstorage.NewMemoryUsersRepo(memDB)
	authUseCase := usecase.NewAuthUseCase(usersRepo, []byte(secret))
	authHandler := adapterhttp.NewAuthHandler(authUseCase)

	postsRepo := localstorage.NewMemoryPostsRepo(memDB)
	postsUseCase := usecase.NewPostsUseCase(postsRepo)
	postsHandler := adapterhttp.NewPostsHandler(ctx, postsUseCase)

	postRepo := localstorage.NewMemoryPostRepo(memDB)
	postUseCase := usecase.NewPostUseCase(postRepo)
	postHandler := adapterhttp.NewPostHandler(ctx, postUseCase)

	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	mux.Handle("/", webHandler)

	auth := middleware.AuthMiddleware([]byte(secret))

	mux.HandleFunc("GET /api/posts", postsHandler.GetAll)
	mux.HandleFunc("GET /api/posts/", postsHandler.GetAll)
	mux.Handle("POST /api/posts", auth(http.HandlerFunc(postsHandler.Create)))
	mux.Handle("POST /api/posts/", auth(http.HandlerFunc(postsHandler.Create)))
	mux.HandleFunc("GET /api/posts/{category}", postsHandler.GetByCategory)

	mux.Handle("/api/post/{postId}", postHandler)

	mux.HandleFunc("POST /api/register", authHandler.Register)
	mux.HandleFunc("POST /api/login", authHandler.Login)

	fmt.Println("server start")
	if err := http.ListenAndServe(":8080", middleware.LogMiddleware(mux)); err != nil {
		fmt.Println("server error:", err)
	}
}
