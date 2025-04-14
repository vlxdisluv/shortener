package config

import (
	"flag"
	"os"
)

type Config struct {
	Environment string // "production" or "development"
	Addr        string
	BaseURL     string
	LogLevel    string
}

var (
	environment string
	addr        string
	baseURL     string
	logLevel    string
)

func init() {
	flag.StringVar(&addr, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&baseURL, "b", "", "Base URL for shortened links")
	flag.StringVar(&logLevel, "l", "info", "Log Level")
	flag.StringVar(&environment, "e", "development", "Environment")
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

	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		logLevel = envLogLevel
	}

	if env := os.Getenv("ENVIRONMENT"); env != "" {
		environment = env
	}

	return &Config{
		Environment: environment,
		Addr:        addr,
		BaseURL:     baseURL,
		LogLevel:    logLevel,
	}
}
