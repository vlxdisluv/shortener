package file

import (
	"context"
	"encoding/json"
	"io"
	"sync"

	"github.com/vlxdisluv/shortener/internal/app/logger"
	"github.com/vlxdisluv/shortener/internal/app/storage"
	"github.com/vlxdisluv/shortener/internal/app/storage/file/internal/filestore"
	"go.uber.org/zap"
)

type ShortURLRepository struct {
	mu        sync.RWMutex
	hashMap   map[string]string
	fileStore *filestore.Store
}

type entry struct {
	Hash string `json:"hash"`
	URL  string `json:"url"`
}

func NewShortURLRepository(path string) (*ShortURLRepository, error) {
	fs, err := filestore.LoadFile(path)
	if err != nil {
		return nil, err
	}

	r := &ShortURLRepository{
		hashMap:   make(map[string]string),
		fileStore: fs,
	}

	for {
		raw, err := fs.ReadRaw()

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		var e entry
		if err := json.Unmarshal(raw, &e); err != nil {
			logger.Log.Warn("skipping invalid entry: %v", zap.Error(err))
			continue
		}

		if e.URL == "" || e.Hash == "" {
			logger.Log.Warn("skipping incomplete entry: %q", zap.Binary("fileRaw", raw))
			continue
		}

		r.hashMap[e.Hash] = e.URL
	}

	return r, nil
}

func (r *ShortURLRepository) Save(_ context.Context, hash, original string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.hashMap[hash]; exists {
		return storage.ErrConflict
	}

	r.hashMap[hash] = original
	if err := r.fileStore.Append(entry{Hash: hash, URL: original}); err != nil {
		return err
	}
	return r.fileStore.Sync()
}

func (r *ShortURLRepository) Get(_ context.Context, hash string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	url, ok := r.hashMap[hash]
	if !ok {
		return "", storage.ErrNotFound
	}
	return url, nil
}

func (r *ShortURLRepository) GetByOriginal(_ context.Context, original string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// TODO can be improved from O(n) to O(1)
	for hash, url := range r.hashMap {
		if url == original {
			return hash, nil
		}
	}
	
	return "", storage.ErrNotFound
}

func (r *ShortURLRepository) Close() error {
	return r.fileStore.Close()
}

// no-op
func (r *ShortURLRepository) WithTx(_ storage.Tx) storage.ShortURLRepository {
	return r
}
