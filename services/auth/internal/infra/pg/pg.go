// Package pg — PostgreSQL + PostGIS infra (pgx v5, 2026)
package pg

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect creates a connection pool. DSN must include sslmode.
func Connect(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	return pgxpool.NewWithConfig(ctx, config)
}

// EnsurePostGIS checks that PostGIS extension is available (required for geolocation).
func EnsurePostGIS(ctx context.Context, pool *pgxpool.Pool) error {
	var exists bool
	err := pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = 'postgis')").Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		_, err = pool.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS postgis")
		return err
	}
	return nil
}
