package server

import (
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/vlxdisluv/shortener/config"
	"github.com/vlxdisluv/shortener/internal/app/handlers"
	"github.com/vlxdisluv/shortener/internal/app/logger"
	customMiddleware "github.com/vlxdisluv/shortener/internal/app/middleware"
	"github.com/vlxdisluv/shortener/internal/app/storage"
	"go.uber.org/zap"
	"net/http"
)

func Start(cfg *config.Config) {
	repo := storage.NewInMemoryURLStore()
	h := handlers.NewShortURLHandler(repo)

	r := chi.NewRouter()

	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.Recoverer)
	r.Use(customMiddleware.RequestLogger)

	r.Post("/", h.CreateShortURLFromRawBody)
	r.Get("/{hash}", h.GetShortURL)
	r.Post("/api/shorten", h.CreateShortURLFromJSON)

	logger.Log.Info("Server started successfully",
		zap.String("address", cfg.Addr),
		zap.String("baseURL", cfg.BaseURL),
		zap.String("logLevel", cfg.LogLevel),
	)

	if err := http.ListenAndServe(cfg.Addr, r); err != nil {
		logger.Log.Fatal("server failed to start", zap.Error(err))
		//log.Fatalf("server failed to start: %v", err)
	}
}
