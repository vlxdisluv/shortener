package server

import (
	"github.com/vlxdisluv/shortener/internal/app/handlers"
	"github.com/vlxdisluv/shortener/internal/app/storage"
	"log"
	"net/http"
)

func Start(addr string) {
	repo := storage.NewInMemoryURLStore()
	shortURLHandler := handlers.NewShortURLHandler(repo)

	mux := http.NewServeMux()
	mux.HandleFunc("/", shortURLHandler)

	log.Printf("Starting server on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
