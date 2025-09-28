package file

import (
	"encoding/json"
	"io"
	"path/filepath"
	"sync"

	"github.com/vlxdisluv/shortener/internal/app/logger"
	"github.com/vlxdisluv/shortener/internal/app/storage/file/internal/filestore"
	"go.uber.org/zap"
)

type CounterRepository struct {
	mu        sync.Mutex
	value     uint64
	fileStore *filestore.Store
}

type counterEntry struct {
	Value uint64 `json:"value"`
}

func NewCounterRepository(path string) (*CounterRepository, error) {
	fs, err := filestore.LoadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	r := &CounterRepository{fileStore: fs}

	for {
		raw, err := fs.ReadRaw()
		if err == io.EOF {
			break
		}
		if err != nil {
			_ = fs.Close()
			return nil, err
		}
		var e counterEntry
		if err := json.Unmarshal(raw, &e); err != nil {
			logger.Log.Warn("counter: skipping invalid entry", zap.Error(err))
			continue
		}
		r.value = e.Value
	}

	return r, nil
}

func (r *CounterRepository) Next() (uint64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.value++
	if err := r.fileStore.Append(counterEntry{Value: r.value}); err != nil {
		r.value--
		return 0, err
	}
	return r.value, nil
}

func (r *CounterRepository) Close() error {
	return r.fileStore.Close()
}
