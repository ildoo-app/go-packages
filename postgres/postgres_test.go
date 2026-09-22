package postgres

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestPingNilDB(t *testing.T) {
	t.Parallel()

	var db *DB
	if err := db.Ping(context.Background()); err == nil {
		t.Fatal("expected error for nil db")
	}

	if err := (&DB{}).Ping(context.Background()); err == nil {
		t.Fatal("expected error for nil embedded pool")
	}
}

func TestNewRejectsInvalidConfig(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := New(ctx, &Config{})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

// TestNewPrewarmsMinConns asserts the pool reaches MinConns idle connections
// after New returns. Requires POSTGRES_TEST_DSN (e.g. postgres://user:pass@localhost:5432/db?sslmode=disable).
func TestNewPrewarmsMinConns(t *testing.T) {
	dsn := os.Getenv("POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("POSTGRES_TEST_DSN not set")
	}

	const minConns int32 = 2
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := New(ctx, &Config{
		DSN:               dsn,
		MinConns:          minConns,
		MaxConns:          4,
		ConnectTimeout:    5 * time.Second,
		HealthCheckPeriod: time.Minute,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer db.Close()

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if db.Stat().IdleConns() >= minConns {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("IdleConns=%d, want >= %d", db.Stat().IdleConns(), minConns)
}
