package main

import (
	"fmt"
	"net/http"

	handler "github.com/inraise/asperitas/internal/adapter/http"
)

func main() {
	mux := http.NewServeMux()

	webHandler := handler.NewWebHandler()

	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	mux.Handle("/{$}", webHandler)

	fmt.Println("server start")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Println("server error:", err)
	}
}
