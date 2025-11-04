package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/vlxdisluv/shortener/internal/app/shortener"
	"github.com/vlxdisluv/shortener/internal/app/storage"
)

type Storage interface {
	ShortURLs() storage.ShortURLRepository
	Counters() storage.CounterRepository
	UnitOfWork() storage.UnitOfWork
}

type ShortURLHandler struct {
	storage Storage
}

func NewShortURLHandler(storage Storage) *ShortURLHandler {
	return &ShortURLHandler{storage: storage}
}

type CreateShortURLReq struct {
	URL string `json:"url"`
}

type CreateShortURLResp struct {
	ShortURL string `json:"result"`
}

type CreateShortURLBatchReq struct {
	OrigURL       string `json:"original_url"`
	CorrelationID string `json:"correlation_id"`
}

type CreateShortURLBatchResp struct {
	ShortURL      string `json:"short_url"`
	CorrelationID string `json:"correlation_id"`
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

	id, err := h.storage.Counters().Next(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	hash := shortener.Generate(id, 7)

	if err := h.storage.ShortURLs().Save(r.Context(), hash, string(body)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	host := r.Host
	shortURL := fmt.Sprintf("http://%s/%s", host, hash)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(shortURL))
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

	id, err := h.storage.Counters().Next(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	hash := shortener.Generate(id, 7)

	if err := h.storage.ShortURLs().Save(r.Context(), hash, shortURLReq.URL); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	host := r.Host
	shortURL := fmt.Sprintf("http://%s/%s", host, hash)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(CreateShortURLResp{ShortURL: shortURL})
}

func (h *ShortURLHandler) GetShortURL(w http.ResponseWriter, r *http.Request) {
	hash := chi.URLParam(r, "hash")

	original, err := h.storage.ShortURLs().Get(r.Context(), hash)
	if err != nil {
		http.Error(w, fmt.Sprintf("short url does not exist for %s", hash), http.StatusNotFound)
		return
	}

	http.Redirect(w, r, original, http.StatusTemporaryRedirect)
}

func (h *ShortURLHandler) CreateShortURLsBatch(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var req []CreateShortURLBatchReq
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if len(req) == 0 {
		http.Error(w, "empty batch", http.StatusBadRequest)
		return
	}
	for i, item := range req {
		if item.OrigURL == "" {
			http.Error(w, fmt.Sprintf("item %d: original_url is required", i), http.StatusBadRequest)
			return
		}
	}

	tx, err := h.storage.UnitOfWork().Begin(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	shortURLRepo := h.storage.ShortURLs().WithTx(tx)
	counterRepo := h.storage.Counters().WithTx(tx)

	var results []CreateShortURLBatchResp
	for _, item := range req {
		id, err := counterRepo.Next(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		hash := shortener.Generate(id, 7)

		if err := shortURLRepo.Save(r.Context(), hash, item.OrigURL); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		host := r.Host
		shortURL := fmt.Sprintf("http://%s/%s", host, hash)

		results = append(results, CreateShortURLBatchResp{CorrelationID: item.CorrelationID, ShortURL: shortURL})
	}
	if err := tx.Commit(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(results)
}
