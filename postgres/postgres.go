package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB is a Postgres handle backed by a pgx connection pool.
// Embedding *pgxpool.Pool keeps Query/Exec/Acquire/etc. available on *DB.
type DB struct {
	*pgxpool.Pool
}

// New creates a pgx connection pool, pings it, and returns the ready handle.
// The caller owns the DB and must Close it on shutdown.
func New(ctx context.Context, cfg *Config) (DBTX, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	cfg.withDefaults()

	connString, err := cfg.connString()
	if err != nil {
		return nil, err
	}

	poolCfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("postgres: parse config: %w", err)
	}

	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MinConns = cfg.MinConns
	poolCfg.MaxConnLifetime = cfg.MaxConnLifetime
	poolCfg.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolCfg.HealthCheckPeriod = cfg.HealthCheckPeriod

	connectCtx, cancel := context.WithTimeout(ctx, cfg.ConnectTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(connectCtx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("postgres: create pool: %w", err)
	}

	db := &DB{Pool: pool}

	if err := db.Ping(connectCtx); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// Ping verifies the pool can reach Postgres within the context deadline.
func (db *DB) Ping(ctx context.Context) error {
	if db == nil || db.Pool == nil {
		return fmt.Errorf("postgres: db is nil")
	}
	if err := db.Pool.Ping(ctx); err != nil {
		return fmt.Errorf("postgres: ping: %w", err)
	}
	return nil
}

// Healthy is a convenience health-check helper with a short timeout.
// Suitable for readiness probes.
func (db *DB) Healthy(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return db.Ping(pingCtx)
}
