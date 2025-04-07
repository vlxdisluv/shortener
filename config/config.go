package config

import (
	"flag"
	"os"
)

type Config struct {
	Addr    string
	BaseURL string
}

var (
	addr    string
	baseURL string
)

func init() {
	flag.StringVar(&addr, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&baseURL, "b", "", "Base URL for shortened links")
}

func Load() *Config {
	if baseURL == "" {
		baseURL = "http://" + addr
	}

	flag.Parse()

	if envAddr := os.Getenv("SERVER_ADDR"); envAddr != "" {
		addr = envAddr
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		baseURL = envBaseURL
	}

	return &Config{
		Addr:    addr,
		BaseURL: baseURL,
	}
}
