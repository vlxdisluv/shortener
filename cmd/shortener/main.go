package main

import (
	"fmt"
	"github.com/vlxdisluv/shortener/config"
	"github.com/vlxdisluv/shortener/internal/app/server"
)

func main() {
	cfg := config.Load()

	fmt.Println("ServerAddress:", cfg.Addr)
	fmt.Println("BaseURL:", cfg.BaseURL)

	server.Start(cfg.Addr)
}
