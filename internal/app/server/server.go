package server

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	storagefactory "github.com/vlxdisluv/shortener/internal/app/storage/factory"

	"github.com/vlxdisluv/shortener/config"
	"github.com/vlxdisluv/shortener/internal/app/handlers"
	"github.com/vlxdisluv/shortener/internal/app/logger"
	customMiddleware "github.com/vlxdisluv/shortener/internal/app/middleware"

	"go.uber.org/zap"
)

func Start(cfg *config.Config) {
	//fmt.Println("db", cfg.DatabaseDSN)
	//poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseDSN)
	//if err != nil {
	//	log.Fatalln("Unable to parse DATABASE_URL:", err)
	//}
	//
	//db, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	//if err != nil {
	//	log.Fatalln("Unable to create connection pool:", err)
	//}

	//repo, err := file.NewShortURLRepository(cfg.FileStoragePath)
	//if err != nil {
	//	logger.Log.Fatal("server failed to init storage", zap.Error(err))
	//}
	//defer repo.Close()
	//
	//counterPath := cfg.FileStoragePath + ".seq"
	//counter, err := file.NewCounterRepository(counterPath)
	//if err != nil {
	//	logger.Log.Fatal("server failed to init counter", zap.Error(err))
	//}
	//defer counter.Close()

	storage, err := storagefactory.New(context.Background(), cfg)
	if err != nil {
		logger.Log.Error("server failed to init storage", zap.Error(err))
		return
	}
	defer storage.Close(context.Background())

	h := handlers.NewShortURLHandler(storage)
	//hh := handlers.NewHealthHandler(storage)

	r := chi.NewRouter()
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.Recoverer)
	r.Use(customMiddleware.RequestLogger)
	r.Use(customMiddleware.GzipCompressor)

	r.Post("/", h.CreateShortURLFromRawBody)
	r.Get("/{hash}", h.GetShortURL)
	r.Post("/api/shorten", h.CreateShortURLFromJSON)
	//r.Get("/ping", hh.DBHealth)

	logger.Log.Info("Server started successfully",
		zap.String("address", cfg.Addr),
		zap.String("baseURL", cfg.BaseURL),
		zap.String("logLevel", cfg.LogLevel),
	)

	if err := http.ListenAndServe(cfg.Addr, r); err != nil {
		logger.Log.Fatal("server failed to start", zap.Error(err))
	}
}
