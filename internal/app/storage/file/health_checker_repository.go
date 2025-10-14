package file

import (
	"context"
	"fmt"
)

type HealthCheckerRepository struct{}

func NewHealthCheckerRepository() (*HealthCheckerRepository, error) {
	return &HealthCheckerRepository{}, nil
}

func (r *HealthCheckerRepository) Ping(_ context.Context) error {
	return fmt.Errorf("not implemented")
}
