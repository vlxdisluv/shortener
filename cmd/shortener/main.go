package main

import (
	"log"

	_ "github.com/joho/godotenv/autoload"
	"github.com/vlxdisluv/shortener/config"
	"github.com/vlxdisluv/shortener/internal/app/logger"
	"github.com/vlxdisluv/shortener/internal/app/server"
)

func main() {
	cfg := config.Load()

	if err := logger.Initialize(cfg); err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Log.Sync()

	server.Start(cfg)
}
