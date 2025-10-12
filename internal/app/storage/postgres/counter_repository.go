package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CounterRepository struct {
	db *pgxpool.Pool
}

func NewCounterRepository(db *pgxpool.Pool) (*CounterRepository, error) {
	return &CounterRepository{db: db}, nil
}

func (r *CounterRepository) Next(ctx context.Context) (uint64, error) {
	const q = `SELECT nextval('url_counter');`

	var id uint64

	err := r.db.QueryRow(ctx, q).Scan(&id)
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
