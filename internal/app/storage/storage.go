package storage

import (
	"context"
	"errors"
)

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

type ShortURLRepository interface {
	Save(ctx context.Context, hash string, original string) error
	Get(ctx context.Context, hash string) (string, error)
	Close() error
}

type CounterRepository interface {
	Next(ctx context.Context) (uint64, error)
	Close() error
}

type HealthCheckRepository interface {
	Ping(ctx context.Context) error
}
