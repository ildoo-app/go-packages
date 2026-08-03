package postgres

import (
	"context"
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
