package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ShortURLRepository struct {
	db *pgxpool.Pool
}

func NewShortURLRepository(db *pgxpool.Pool) (*ShortURLRepository, error) {
	return &ShortURLRepository{db: db}, nil
}

func (r *ShortURLRepository) Save(context context.Context, hash, original string) error {
	const q = `INSERT INTO short_urls(hash, original) VALUES ($1, $2) ON CONFLICT (hash) DO NOTHING`
	_, err := r.db.Exec(context, q, hash, original)
	if err != nil {
		return err
	}
	return nil
}
