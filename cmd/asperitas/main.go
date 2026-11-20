package main

import (
	"context"
	"fmt"
	"net/http"

	adapterhttp "github.com/inraise/asperitas/internal/adapter/http"
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

	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	mux.Handle("/", webHandler)
	mux.Handle("/api/posts", postsHandler)
	mux.Handle("/api/posts/", postsHandler)

	fmt.Println("server start")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Println("server error:", err)
	}
}
