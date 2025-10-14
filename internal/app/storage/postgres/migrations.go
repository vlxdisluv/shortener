package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	mgpg "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
	appmigrations "github.com/vlxdisluv/shortener"
)

func RunMigrations(_ context.Context, dsn string) error {
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open sql db: %w", err)
	}
	defer sqlDB.Close()

	driver, err := mgpg.WithInstance(sqlDB, &mgpg.Config{})
	if err != nil {
		return fmt.Errorf("migrate db driver: %w", err)
	}

	src, err := iofs.New(appmigrations.FS, "db/migrations")
	if err != nil {
		return fmt.Errorf("migrate iofs: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "postgres", driver)
	if err != nil {
		return fmt.Errorf("migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}
