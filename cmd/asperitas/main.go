package main

import (
	"context"
	"fmt"
	"net/http"

	adapterhttp "github.com/inraise/asperitas/internal/adapter/http"
	"github.com/inraise/asperitas/internal/adapter/repository/localstorage"
	"github.com/inraise/asperitas/internal/usecase"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mux := http.NewServeMux()

	webHandler := adapterhttp.NewWebHandler()

	tasksRepo := localstorage.NewMemoryTaskRepo()
	tasksUseCase := usecase.NewTaskUseCase(tasksRepo)
	tasksHandler := adapterhttp.NewTaskHandler(ctx, tasksUseCase)

	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	mux.Handle("/", webHandler)
	mux.Handle("/api/posts", tasksHandler)
	mux.Handle("/api/posts/", tasksHandler)

	fmt.Println("server start")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Println("server error:", err)
	}
}
