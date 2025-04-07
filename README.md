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
4. **Run the Application:**
   ```bash
   go run cmd/shortener/main.go
   ```
   The application will start on the default port (e.g., `8080`). Adjust the configuration as needed.

## Testing
To run the automated tests, use:
```bash
go test ./...
```