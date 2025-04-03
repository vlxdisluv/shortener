package config

import "flag"

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

	return &Config{
		Addr:    addr,
		BaseURL: baseURL,
	}
}
