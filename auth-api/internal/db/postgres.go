// Package db provides the PostgreSQL connection pool and migration runner.
package db

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/001_init.sql
var migration001 string

//go:embed migrations/002_verification_token_expires_at.sql
var migration002 string

// Connect creates a pgx connection pool and runs migrations.
func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.ParseConfig: %w", err)
	}
	poolCfg.MaxConns = 20
	poolCfg.MinConns = 2

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.NewWithConfig: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}

	for i, sql := range []string{migration001, migration002} {
		if _, err := pool.Exec(ctx, sql); err != nil {
			pool.Close()
			return nil, fmt.Errorf("migration %d: %w", i+1, err)
		}
	}

	return pool, nil
}
