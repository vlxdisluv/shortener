# Go Shortener

## Overview
Go Shortener is a simple URL shortening service built with Go. It allows you to shorten long URLs, redirect users to the original URLs, and optionally collect click statistics. This project is intended for both learning purposes and practical use.

## Features
- **URL Shortening:** Generate short links from long URLs.
- **Redirection:** Automatically redirect from a short URL to the original long URL.
- **REST API:** Easily integrate with other services via a well-defined API.

## Technologies
- **Language:** Go

## Installation and Setup
1. **Clone the Repository:**
   ```bash
   git clone https://github.com/vlxdisluv/shortener.git
   ```
2. **Navigate to the Project Directory:**
   ```bash
   cd shortener
   ```
3. **Install Dependencies:**
   ```bash
   go mod tidy
   ```

## Environment Configuration
- A sample environment file is provided as `.env.example`.
- Copy it to `.env` and adjust values as needed:
  ```bash
  cp .env.example .env
  ```
- The application autoloads `.env` at startup.

Supported environment variables:
- `SERVER_ADDR` — HTTP server address (e.g., `localhost:8080`).
- `BASE_URL` — Base URL for generated short links (falls back to `http://<SERVER_ADDR>`).
- `LOG_LEVEL` — Logging level (e.g., `info`).
- `ENVIRONMENT` — Environment name (e.g., `development`, `production`).
- `FILE_STORAGE_PATH` — Path to JSON file used for file-based storage (default: `/tmp/short-url-db.json`).
- `DATABASE_DSN` — Postgres connection string (enables Postgres storage when set).

## Storage Backends
- By default, the service uses a file-based storage.
- To use Postgres instead of the file storage, provide a Postgres DSN via either:
  - Environment: set `DATABASE_DSN` in `.env` (or the environment), or
  - CLI flag: `-d "<postgres_dsn>"`.
- If neither the flag nor the env var is provided, the file storage will be used.

## Using Docker Compose (Postgres)
A Docker Compose configuration is included to run Postgres locally.

1. Start Postgres and run migrations:
   ```bash
   docker compose up -d
   ```
   This will start a Postgres 15 container and apply migrations.

2. To connect the application to this Postgres instance, set the `DATABASE_DSN` or use the `-d` flag as described in the Running section below.

## Running
- File storage (default):
  ```bash
  go run cmd/shortener/main.go
  ```
- Postgres via env (`.env` or shell):
  ```bash
  export DATABASE_DSN="postgres://shortener:shortener@localhost:5432/postgres?sslmode=disable"
  go run cmd/shortener/main.go
  ```
- Postgres via flag:
  ```bash
  go run cmd/shortener/main.go -d "postgres://shortener:shortener@localhost:5432/postgres?sslmode=disable"
  ```

## Testing
To run the automated tests, use:
```bash
go test ./...
```