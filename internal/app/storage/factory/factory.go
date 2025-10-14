package factory

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vlxdisluv/shortener/config"
	"github.com/vlxdisluv/shortener/internal/app/logger"
	"github.com/vlxdisluv/shortener/internal/app/storage"
	"github.com/vlxdisluv/shortener/internal/app/storage/file"
	"github.com/vlxdisluv/shortener/internal/app/storage/postgres"
	"go.uber.org/zap"
)

type Storage struct {
	short   storage.ShortURLRepository
	counter storage.CounterRepository
	hc      storage.HealthCheckRepository
	closer  func(context.Context)
}

func New(ctx context.Context, cfg *config.Config) (*Storage, error) {
	if cfg.DatabaseDSN != "" {
		if err := postgres.RunMigrations(ctx, cfg.DatabaseDSN); err != nil {
			return nil, fmt.Errorf("run migrations: %w", err)
		}
		
		poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseDSN)
		if err != nil {
			return nil, fmt.Errorf("parse pg: %w", err)
		}

		db, err := pgxpool.NewWithConfig(ctx, poolConfig)
		if err != nil {
			return nil, fmt.Errorf("create pg: %w", err)
		}

		short, err := postgres.NewShortURLRepository(db)
		if err != nil {
			logger.Log.Error("server failed to init pg short url repository", zap.Error(err))
			return nil, fmt.Errorf("create pg short url repo: %w", err)
		}

		counter, err := postgres.NewCounterRepository(db)
		if err != nil {
			logger.Log.Error("server failed to init pg counter repository", zap.Error(err))
			return nil, fmt.Errorf("create pg counter repo: %w", err)
		}

		hc, err := postgres.NewHealthCheckerRepository(db)
		if err != nil {
			logger.Log.Error("server failed to init pg health checker repository", zap.Error(err))
			return nil, fmt.Errorf("create pg health checker repo: %w", err)
		}

		return &Storage{
			short:   short,
			counter: counter,
			hc:      hc,
			closer:  func(context.Context) { db.Close() },
		}, nil
	}

	short, err := file.NewShortURLRepository(cfg.FileStoragePath)
	if err != nil {
		logger.Log.Error("server failed to init file short url repository", zap.Error(err))
		return nil, fmt.Errorf("create file short url repo: %w", err)
	}

	counterPath := cfg.FileStoragePath + ".seq"
	counter, err := file.NewCounterRepository(counterPath)
	if err != nil {
		logger.Log.Error("server failed to init file counter repository", zap.Error(err))
		return nil, fmt.Errorf("create file counter repo: %w", err)
	}

	hc, err := file.NewHealthCheckerRepository()
	if err != nil {
		logger.Log.Error("server failed to init file health checker repository", zap.Error(err))
		return nil, fmt.Errorf("create file health checker repo: %w", err)
	}

	return &Storage{
		short:   short,
		counter: counter,
		hc:      hc,
		closer: func(context.Context) {
			if err := short.Close(); err != nil {
				logger.Log.Warn("file short repo close failed", zap.Error(err))
			}
			if err := counter.Close(); err != nil {
				logger.Log.Warn("file counter repo close failed", zap.Error(err))
			}
		},
	}, nil
}

func (s *Storage) ShortURLs() storage.ShortURLRepository { return s.short }

func (s *Storage) Counters() storage.CounterRepository { return s.counter }

func (s *Storage) HealthCheck() storage.HealthCheckRepository { return s.hc }

func (s *Storage) Close(ctx context.Context) {
	if s.closer != nil {
		s.closer(ctx)
	}
}
