package postgres

import (
	"errors"
	"fmt"
	"net/url"
	"time"
)

const (
	defaultPort              = 5432
	defaultSSLMode           = "prefer"
	defaultMaxConns          = int32(10)
	defaultMinConns          = int32(1)
	defaultMaxConnLifetime   = time.Hour
	defaultMaxConnIdleTime   = 30 * time.Minute
	defaultHealthCheckPeriod = time.Minute
	defaultConnectTimeout    = 5 * time.Second
)

// Config holds Postgres connection and pool settings.
// Callers map their own env vars or secrets into this struct.
type Config struct {
	// DSN is a full connection string (postgres://...). When set, it takes
	// precedence over the discrete fields below.
	DSN string

	Host     string
	Port     uint16
	User     string
	Password string
	Database string
	// SSLMode is used when building a DSN from discrete fields.
	// Defaults to "prefer".
	SSLMode string

	// Pool tuning. Zero values receive sensible defaults in New.
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration

	// ConnectTimeout bounds the initial pool creation / first ping.
	ConnectTimeout time.Duration
}

func (c *Config) validate() error {
	if c == nil {
		return errors.New("postgres: config is nil")
	}
	if c.DSN != "" {
		return nil
	}
	if c.Host == "" {
		return errors.New("postgres: DSN or Host is required")
	}
	if c.User == "" {
		return errors.New("postgres: User is required when DSN is empty")
	}
	if c.Database == "" {
		return errors.New("postgres: Database is required when DSN is empty")
	}
	return nil
}

func (c *Config) withDefaults() {
	if c.Port == 0 {
		c.Port = defaultPort
	}
	if c.SSLMode == "" {
		c.SSLMode = defaultSSLMode
	}
	if c.MaxConns == 0 {
		c.MaxConns = defaultMaxConns
	}
	if c.MinConns == 0 {
		c.MinConns = defaultMinConns
	}
	if c.MaxConnLifetime == 0 {
		c.MaxConnLifetime = defaultMaxConnLifetime
	}
	if c.MaxConnIdleTime == 0 {
		c.MaxConnIdleTime = defaultMaxConnIdleTime
	}
	if c.HealthCheckPeriod == 0 {
		c.HealthCheckPeriod = defaultHealthCheckPeriod
	}
	if c.ConnectTimeout == 0 {
		c.ConnectTimeout = defaultConnectTimeout
	}
}

func (c *Config) connString() (string, error) {
	if c.DSN != "" {
		return c.DSN, nil
	}

	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.User, c.Password),
		Host:   fmt.Sprintf("%s:%d", c.Host, c.Port),
		Path:   "/" + c.Database,
	}
	q := u.Query()
	q.Set("sslmode", c.SSLMode)
	u.RawQuery = q.Encode()
	return u.String(), nil
}
