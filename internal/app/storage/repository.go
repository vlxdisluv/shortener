package storage

import "errors"

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

type ShortURLRepository interface {
	Save(hash string, original string) error
	Get(hash string) (string, error)
	Close() error
}

type CounterRepository interface {
	Next() (uint64, error)
	Close() error
}
