package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vlxdisluv/shortener/internal/app/storage"
)

type ShortURLRepository struct {
	ex   storage.Execer
	pool *pgxpool.Pool
}

func NewShortURLRepository(pool *pgxpool.Pool) (*ShortURLRepository, error) {
	return &ShortURLRepository{ex: pool, pool: pool}, nil
}

func (r *ShortURLRepository) WithTx(tx storage.Tx) storage.ShortURLRepository {
	if tx == nil {
		return r
	}

	if tw, ok := tx.(txWrapper); ok {
		return &ShortURLRepository{ex: tw.tx, pool: r.pool}
	}

	if ex, ok := tx.(storage.Execer); ok {
		return &ShortURLRepository{ex: ex, pool: r.pool}
	}

	return r
}

func (r *ShortURLRepository) Save(ctx context.Context, hash string, original string) error {
	const q = `INSERT INTO short_urls(hash, original) VALUES ($1, $2) ON CONFLICT (hash) DO NOTHING`
	tag, err := r.ex.Exec(ctx, q, hash, original)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return storage.ErrConflict
	}
	return nil
}

func (r *ShortURLRepository) Get(ctx context.Context, hash string) (string, error) {
	const q = `SELECT original FROM short_urls WHERE hash = $1`
	var original string
	if err := r.ex.QueryRow(ctx, q, hash).Scan(&original); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", storage.ErrNotFound
		}
		return "", err
	}
	return original, nil
}

// Close implements the Repository interface.
// For the Postgres repository this is a no-op, because the repository
// does not own the database connection pool. The pool must be closed
// by the application (e.g. in main), not by the repository itself.
func (r *ShortURLRepository) Close() error {
	return nil
}
