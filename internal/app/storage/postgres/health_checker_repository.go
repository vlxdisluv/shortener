package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthCheckerRepository struct {
	db *pgxpool.Pool
}

func NewHealthCheckerRepository(db *pgxpool.Pool) (*HealthCheckerRepository, error) {
	return &HealthCheckerRepository{db: db}, nil
}

func (r *HealthCheckerRepository) Ping(ctx context.Context) error {
	err := r.db.Ping(ctx)

	if err != nil {
		return err
	}

	return nil
}
