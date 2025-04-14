package main

import (
	"github.com/vlxdisluv/shortener/config"
	"github.com/vlxdisluv/shortener/internal/app/logger"
	"github.com/vlxdisluv/shortener/internal/app/server"
	"log"
)

func main() {
	cfg := config.Load()

	if err := logger.Initialize(cfg); err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Log.Sync()

	server.Start(cfg)
}
