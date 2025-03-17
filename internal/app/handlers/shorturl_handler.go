package handlers

import (
	"fmt"
	"github.com/vlxdisluv/shortener/internal/app/storage"
	"io"
	"net/http"
)

type ShortURLHandler struct {
	repo storage.URLRepository
}

func NewShortURLHandler(repo storage.URLRepository) http.HandlerFunc {
	h := &ShortURLHandler{repo: repo}
	return h.handle
}

func (h *ShortURLHandler) handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getShortURL(w, r)
	case http.MethodPost:
		h.createShortURL(w, r)
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, fmt.Sprintf("Method %s not allowed", r.Method), http.StatusMethodNotAllowed)
	}
}

func (h *ShortURLHandler) createShortURL(w http.ResponseWriter, r *http.Request) {
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

	hash := "EwHXdJfB"

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

func (h *ShortURLHandler) getShortURL(w http.ResponseWriter, r *http.Request) {
	hash := r.URL.Path[1:]

	original, err := h.repo.Get(hash)
	if err != nil {
		http.Error(w, fmt.Sprintf("short url does not exist for %s", hash), http.StatusNotFound)
		return
	}

	http.Redirect(w, r, original, http.StatusTemporaryRedirect)
}
