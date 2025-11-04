package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vlxdisluv/shortener/internal/app/storage"
)

type CounterRepository struct {
	ex   storage.Execer
	pool *pgxpool.Pool
}

func NewCounterRepository(pool *pgxpool.Pool) (*CounterRepository, error) {
	return &CounterRepository{pool: pool, ex: pool}, nil
}

func (r *CounterRepository) WithTx(tx storage.Tx) storage.CounterRepository {
	if tx == nil {
		return r
	}

	if tw, ok := tx.(txWrapper); ok {
		return &CounterRepository{ex: tw.tx, pool: r.pool}
	}

	if ex, ok := tx.(storage.Execer); ok {
		return &CounterRepository{ex: ex, pool: r.pool}
	}

	return r
}

func (r *CounterRepository) Next(ctx context.Context) (uint64, error) {
	const q = `SELECT nextval('url_counter');`

	var id uint64

	err := r.ex.QueryRow(ctx, q).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// Close implements the Repository interface.
// For the Postgres repository this is a no-op, because the repository
// does not own the database connection pool. The pool must be closed
// by the application (e.g. in main), not by the repository itself.
func (r *CounterRepository) Close() error {
	return nil
}
