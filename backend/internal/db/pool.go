package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool creates a pgx connection pool and waits (up to 30s) for Postgres to
// accept connections. The retry loop matters for `docker-compose up`: the api
// container may race postgres readiness.
func NewPool(ctx context.Context, databaseURL string, maxConns, minConns int32) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	poolCfg.MaxConns = maxConns
	poolCfg.MinConns = minConns

	deadline := time.Now().Add(30 * time.Second)
	for {
		pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
		if err == nil {
			if pingErr := pool.Ping(ctx); pingErr == nil {
				return pool, nil
			} else {
				pool.Close()
				err = pingErr
			}
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("database not reachable within 30s: %w", err)
		}
		time.Sleep(500 * time.Millisecond)
	}
}
