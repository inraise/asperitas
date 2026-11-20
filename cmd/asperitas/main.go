package main

import (
	"context"
	"fmt"
	"net/http"

	adapterhttp "github.com/inraise/asperitas/internal/adapter/handler/http"
	"github.com/inraise/asperitas/internal/adapter/repository/localstorage"
	"github.com/inraise/asperitas/internal/storage"
	"github.com/inraise/asperitas/internal/usecase"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	memDB := storage.NewStorage()
	mux := http.NewServeMux()

	webHandler := adapterhttp.NewWebHandler()

	postsRepo := localstorage.NewMemoryPostsRepo(memDB)
	postsUseCase := usecase.NewPostsUseCase(postsRepo)
	postsHandler := adapterhttp.NewPostsHandler(ctx, postsUseCase)

	postRepo := localstorage.NewMemoryPostRepo(memDB)
	postUseCase := usecase.NewPostUseCase(postRepo)
	postHandler := adapterhttp.NewPostHandler(ctx, postUseCase)

	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	mux.Handle("/", webHandler)

	mux.HandleFunc("/api/posts", postsHandler.GetAll)
	mux.HandleFunc("/api/posts/", postsHandler.GetAll)
	mux.HandleFunc("/api/posts/{category}", postsHandler.GetByCategory)

	mux.Handle("/api/post/{postId}", postHandler)

	fmt.Println("server start")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Println("server error:", err)
	}
}
