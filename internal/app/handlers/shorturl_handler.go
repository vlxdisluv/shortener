package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/vlxdisluv/shortener/internal/app/shortener"
	"github.com/vlxdisluv/shortener/internal/app/storage"
	"io"
	"net/http"
)

type ShortURLHandler struct {
	repo storage.URLRepository
}

func NewShortURLHandler(repo storage.URLRepository) *ShortURLHandler {
	return &ShortURLHandler{repo: repo}
}

type CreateShortURLReq struct {
	URL string `json:"url"`
}

type CreateShortURLResp struct {
	ShortURL string `json:"result"`
}

func (h *ShortURLHandler) CreateShortURLFromRawBody(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	}

	id := h.repo.NextID()
	hash := shortener.Generate(id, 7)

	if err := h.repo.Save(hash, string(body)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	host := r.Host
	shortURL := fmt.Sprintf("http://%s/%s", host, hash)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func (h *ShortURLHandler) CreateShortURLFromJSON(w http.ResponseWriter, r *http.Request) {
	var shortURLReq CreateShortURLReq

	if err := json.NewDecoder(r.Body).Decode(&shortURLReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if shortURLReq.URL == "" {
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}

	id := h.repo.NextID()
	hash := shortener.Generate(id, 7)

	if err := h.repo.Save(hash, shortURLReq.URL); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	host := r.Host
	shortURL := fmt.Sprintf("http://%s/%s", host, hash)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(CreateShortURLResp{ShortURL: shortURL})
}

func (h *ShortURLHandler) GetShortURL(w http.ResponseWriter, r *http.Request) {
	hash := chi.URLParam(r, "hash")

	original, err := h.repo.Get(hash)
	if err != nil {
		http.Error(w, fmt.Sprintf("short url does not exist for %s", hash), http.StatusNotFound)
		return
	}

	http.Redirect(w, r, original, http.StatusTemporaryRedirect)
}
