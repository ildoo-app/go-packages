package redis

import (
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const (
	defaultPort            = 6379
	defaultPoolSize        = 10
	defaultMinIdleConns    = 1
	defaultConnMaxLifetime = time.Hour
	defaultConnMaxIdleTime = 30 * time.Minute
	defaultConnectTimeout  = 5 * time.Second
)

// Config holds Redis connection and pool settings.
// Callers map their own env vars or secrets into this struct.
type Config struct {
	// DSN is a full connection string (redis://... or rediss://...). When set,
	// it takes precedence over the discrete fields below.
	DSN string

	Host     string
	Port     uint16
	Username string
	Password string
	// DB is the Redis logical database index (0-15 typically).
	DB int

	// Pool tuning. Zero values receive sensible defaults in New.
	PoolSize        int
	MinIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration

	// ConnectTimeout bounds the initial client creation / first ping.
	ConnectTimeout time.Duration
}

func (c *Config) validate() error {
	if c == nil {
		return errors.New("redis: config is nil")
	}
	if c.DSN != "" {
		return nil
	}
	if c.Host == "" {
		return errors.New("redis: DSN or Host is required")
	}
	return nil
}

func (c *Config) withDefaults() {
	if c.Port == 0 {
		c.Port = defaultPort
	}
	if c.PoolSize == 0 {
		c.PoolSize = defaultPoolSize
	}
	if c.MinIdleConns == 0 {
		c.MinIdleConns = defaultMinIdleConns
	}
	if c.ConnMaxLifetime == 0 {
		c.ConnMaxLifetime = defaultConnMaxLifetime
	}
	if c.ConnMaxIdleTime == 0 {
		c.ConnMaxIdleTime = defaultConnMaxIdleTime
	}
	if c.ConnectTimeout == 0 {
		c.ConnectTimeout = defaultConnectTimeout
	}
}

func (c *Config) addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func (c *Config) options() (*goredis.Options, error) {
	var err error

	opts := &goredis.Options{
		Addr:     c.addr(),
		Username: c.Username,
		Password: c.Password,
		DB:       c.DB,
	}

	if c.DSN != "" {
		if opts, err = goredis.ParseURL(c.DSN); err != nil {
			return nil, fmt.Errorf("redis: parse DSN: %w", err)
		}
	}

	applyPoolSettings(opts, c)
	return opts, nil
}

func applyPoolSettings(opts *goredis.Options, c *Config) {
	opts.PoolSize = c.PoolSize
	opts.MinIdleConns = c.MinIdleConns
	opts.ConnMaxLifetime = c.ConnMaxLifetime
	opts.ConnMaxIdleTime = c.ConnMaxIdleTime
	opts.DialTimeout = c.ConnectTimeout
}
