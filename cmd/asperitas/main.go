package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/inraise/asperitas/internal/controller/http/middleware"
	adapterhttp "github.com/inraise/asperitas/internal/controller/http/v1"
	adapterweb "github.com/inraise/asperitas/internal/controller/http/web"
	"github.com/inraise/asperitas/internal/repository/localstorage"
	memorystorage "github.com/inraise/asperitas/internal/repository/localstorage/storage"
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

	webHandler := adapterweb.NewWebHandler()

	usersRepo := localstorage.NewMemoryUsersRepo(memDB)
	authUseCase := usecase.NewAuthUseCase(usersRepo, []byte(secret))
	authHandler := adapterhttp.NewAuthHandler(authUseCase)

	postsRepo := localstorage.NewMemoryPostsRepo(memDB)
	postsUseCase := usecase.NewPostsUseCase(postsRepo)
	postsHandler := adapterhttp.NewPostsHandler(ctx, postsUseCase)

	postRepo := localstorage.NewMemoryPostRepo(memDB)
	postUseCase := usecase.NewPostUseCase(postRepo)
	postHandler := adapterhttp.NewPostHandler(ctx, postUseCase)

	userRepo := localstorage.NewMemoryUserRepo(memDB)
	userUseCase := usecase.NewUserUseCase(userRepo)
	userHandler := adapterhttp.NewUserHandler(ctx, userUseCase)

	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	mux.Handle("/{$}", webHandler)

	// Аутентификация
	auth := middleware.AuthMiddleware([]byte(secret))
	mux.HandleFunc("POST /api/register", authHandler.Register)
	mux.HandleFunc("POST /api/login", authHandler.Login)

	// Посты
	mux.HandleFunc("GET /api/posts", postsHandler.GetAll)
	mux.HandleFunc("GET /api/posts/", postsHandler.GetAll)
	mux.Handle("POST /api/posts", auth(http.HandlerFunc(postsHandler.Create)))
	mux.Handle("POST /api/posts/", auth(http.HandlerFunc(postsHandler.Create)))

	mux.HandleFunc("GET /api/posts/{category}", postsHandler.GetByCategory)
	mux.HandleFunc("GET /api/user/{username}", userHandler.GetUserPosts)
	mux.HandleFunc("/api/post/{postId}", postHandler.GetPost)
	//

	// Комментарии

	// Голоса

	fmt.Println("server start")
	if err := http.ListenAndServe(":8080", middleware.LogMiddleware(mux)); err != nil {
		fmt.Println("server error:", err)
	}
}
