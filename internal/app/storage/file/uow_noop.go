package file

import (
	"context"

	"github.com/vlxdisluv/shortener/internal/app/storage"
)

type noopTx struct{}

func (noopTx) Commit(ctx context.Context) error   { return nil }
func (noopTx) Rollback(ctx context.Context) error { return nil }

type noopUoW struct{}

func (noopUoW) Begin(ctx context.Context) (storage.Tx, error) { return noopTx{}, nil }

func NewNoopUnitOfWork() storage.UnitOfWork { return noopUoW{} }
