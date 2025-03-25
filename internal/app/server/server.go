package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/vlxdisluv/shortener/internal/app/handlers"
	"github.com/vlxdisluv/shortener/internal/app/storage"
	"log"
	"net/http"
)

func Start(addr string) {
	repo := storage.NewInMemoryURLStore()
	shortURLHandler := handlers.NewShortURLHandler(repo)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/", shortURLHandler)
	r.Get("/{hash}", shortURLHandler)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
