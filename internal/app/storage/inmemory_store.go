package storage

import (
	"encoding/json"
	"fmt"
	"github.com/vlxdisluv/shortener/internal/app/logger"
	"go.uber.org/zap"
	"io"
	"sync"
	"sync/atomic"
)

type InMemoryURLStore struct {
	mu        sync.RWMutex
	hashMap   map[string]string
	seq       int64
	fileStore *FileStore
}

type entry struct {
	Hash string `json:"hash"`
	URL  string `json:"url"`
}

func NewInMemoryURLStore(path string) (*InMemoryURLStore, error) {
	fileStore, err := LoadFile(path)
	if err != nil {
		return nil, err
	}

	memStore := &InMemoryURLStore{
		hashMap:   make(map[string]string),
		fileStore: fileStore,
	}

	for {
		raw, err := fileStore.ReadRaw()

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

		memStore.hashMap[e.Hash] = e.URL
	}

	memStore.seq = int64(len(memStore.hashMap))
	return memStore, nil
}

func (s *InMemoryURLStore) NextID() int64 {
	return atomic.AddInt64(&s.seq, 1)
}

func (s *InMemoryURLStore) Save(hash, original string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	//if _, exists := s.hashMap[hash]; exists {
	//	return fmt.Errorf("hash already exists")
	//}

	s.hashMap[hash] = original

	if err := s.fileStore.Append(entry{Hash: hash, URL: original}); err != nil {
		return err
	}

	return nil
}

func (s *InMemoryURLStore) Get(hash string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, exists := s.hashMap[hash]

	if !exists {
		return "", fmt.Errorf("hash does not exist")
	}

	return url, nil
}
