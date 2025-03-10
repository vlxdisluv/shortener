package storage

import (
	"fmt"
	"sync"
)

type InMemoryURLStore struct {
	mu      sync.RWMutex
	hashMap map[string]string
}

func NewInMemoryURLStore() *InMemoryURLStore {
	return &InMemoryURLStore{
		hashMap: make(map[string]string),
	}
}

func (s *InMemoryURLStore) Save(hash, original string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.hashMap[hash]; exists {
		return fmt.Errorf("hash already exists")
	}

	s.hashMap[hash] = original
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
