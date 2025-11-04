package storage

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

type BatchURL struct {
	CorrelationID string
	URL           string
}

type Execer interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type ShortURLRepository interface {
	Save(ctx context.Context, hash string, original string) error
	Get(ctx context.Context, hash string) (string, error)
	Close() error
	WithTx(tx Tx) ShortURLRepository
}

type CounterRepository interface {
	Next(ctx context.Context) (uint64, error)
	Close() error
	WithTx(tx Tx) CounterRepository
}

type HealthCheckRepository interface {
	Ping(ctx context.Context) error
}

type Tx interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type UnitOfWork interface {
	Begin(ctx context.Context) (Tx, error)
}
