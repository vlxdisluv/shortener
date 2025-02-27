package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
)

type URLStore struct {
	mu      sync.RWMutex
	hashMap map[string]string
}

func NewURLStore() *URLStore {
	return &URLStore{
		hashMap: make(map[string]string),
	}
}

var store = NewURLStore()

func (s *URLStore) Save(hash, original string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hashMap[hash] = original
}

func (s *URLStore) Get(hash string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	url, exists := s.hashMap[hash]
	return url, exists
}

func generateHash() string {
	// TODO: extend
	return "EwHXdJfB"
}

func shortURLHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getShortURL(w, r)
	case http.MethodPost:
		createShortURL(w, r)
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, fmt.Sprintf("Method %s not allowed", r.Method), http.StatusMethodNotAllowed)
	}
}

func createShortURL(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Println("Received data:", string(body))

	host := r.Host
	domain, port, err := net.SplitHostPort(host)
	if err != nil {
		domain = host
		port = "8080"
	}

	hash := generateHash()

	_, exists := store.Get(hash)
	if exists {
		// TODO: fix to conflict error
		http.Error(w, "short url already exists, try again later", http.StatusConflict)
		return
	}

	store.Save(hash, string(body))

	shortURL := fmt.Sprintf("http://%s:%s/%s", domain, port, hash)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func getShortURL(w http.ResponseWriter, r *http.Request) {
	hash := r.URL.String()[1:]

	url, exists := store.Get(hash)
	if !exists {
		http.Error(w, fmt.Sprintf("short url does not exist for %s", hash), http.StatusBadRequest)
		return
	}

	fmt.Sprintln(url)
	//w.Header().Set("Content-Type", "text/plain")
	//w.Write([]byte(url))
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", shortURLHandler)

	port := ":8080"
	log.Printf("Starting server on port %s", port)

	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
