package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Client is a Redis handle backed by go-redis.
// Embedding *goredis.Client keeps Get/Set/Del/Pipeline/etc. available on *Client.
type Client struct {
	*goredis.Client
}

// New creates a go-redis client, pings it, and returns the ready handle.
// The caller owns the Client and must Close it on shutdown.
func New(ctx context.Context, cfg *Config) (*Client, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	cfg.withDefaults()

	opts, err := cfg.options()
	if err != nil {
		return nil, err
	}

	rdb := goredis.NewClient(opts)
	client := &Client{Client: rdb}

	connectCtx, cancel := context.WithTimeout(ctx, cfg.ConnectTimeout)
	defer cancel()

	if err := client.Ping(connectCtx); err != nil {
		_ = client.Close()
		return nil, err
	}

	return client, nil
}

// Ping verifies the client can reach Redis within the context deadline.
func (c *Client) Ping(ctx context.Context) error {
	if c == nil || c.Client == nil {
		return fmt.Errorf("redis: client is nil")
	}
	if err := c.Client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis: ping: %w", err)
	}
	return nil
}

// Healthy is a convenience health-check helper with a short timeout.
// Suitable for readiness probes.
func (c *Client) Healthy(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return c.Ping(pingCtx)
}
