package redis

import (
	"context"
	"testing"
	"time"
)

func TestPingNilClient(t *testing.T) {
	t.Parallel()

	var client *Client
	if err := client.Ping(context.Background()); err == nil {
		t.Fatal("expected error for nil client")
	}

	if err := (&Client{}).Ping(context.Background()); err == nil {
		t.Fatal("expected error for nil embedded client")
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
