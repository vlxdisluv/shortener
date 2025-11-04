package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vlxdisluv/shortener/internal/app/storage"
)

type txWrapper struct{ tx pgx.Tx }

func (t txWrapper) Commit(ctx context.Context) error   { return t.tx.Commit(ctx) }
func (t txWrapper) Rollback(ctx context.Context) error { return t.tx.Rollback(ctx) }

type unitOfWork struct {
	pool *pgxpool.Pool
}

func NewUnitOfWork(pool *pgxpool.Pool) storage.UnitOfWork {
	return &unitOfWork{pool: pool}
}

func (u *unitOfWork) Begin(ctx context.Context) (storage.Tx, error) {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}

	return txWrapper{tx: tx}, nil
}
