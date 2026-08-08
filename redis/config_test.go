package redis

import (
	"strings"
	"testing"
	"time"
)

func TestConfigValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     *Config
		wantErr string
	}{
		{
			name: "dsn is enough",
			cfg:  &Config{DSN: "redis://localhost:6379/0"},
		},
		{
			name: "discrete fields",
			cfg: &Config{
				Host: "localhost",
			},
		},
		{
			name:    "nil config",
			cfg:     nil,
			wantErr: "config is nil",
		},
		{
			name:    "missing everything",
			cfg:     &Config{},
			wantErr: "DSN or Host is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.cfg.validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("got %v, want error containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestConfigWithDefaults(t *testing.T) {
	t.Parallel()

	cfg := &Config{DSN: "redis://localhost/0"}
	cfg.withDefaults()

	if cfg.Port != defaultPort {
		t.Errorf("Port = %d, want %d", cfg.Port, defaultPort)
	}
	if cfg.PoolSize != defaultPoolSize {
		t.Errorf("PoolSize = %d, want %d", cfg.PoolSize, defaultPoolSize)
	}
	if cfg.MinIdleConns != defaultMinIdleConns {
		t.Errorf("MinIdleConns = %d, want %d", cfg.MinIdleConns, defaultMinIdleConns)
	}
	if cfg.ConnMaxLifetime != defaultConnMaxLifetime {
		t.Errorf("ConnMaxLifetime = %v, want %v", cfg.ConnMaxLifetime, defaultConnMaxLifetime)
	}
	if cfg.ConnMaxIdleTime != defaultConnMaxIdleTime {
		t.Errorf("ConnMaxIdleTime = %v, want %v", cfg.ConnMaxIdleTime, defaultConnMaxIdleTime)
	}
	if cfg.ConnectTimeout != defaultConnectTimeout {
		t.Errorf("ConnectTimeout = %v, want %v", cfg.ConnectTimeout, defaultConnectTimeout)
	}
}

func TestConfigWithDefaultsPreservesOverrides(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		DSN:             "redis://localhost/0",
		PoolSize:        25,
		MinIdleConns:    5,
		ConnMaxLifetime: 2 * time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
		ConnectTimeout:  10 * time.Second,
		Port:            6380,
	}
	cfg.withDefaults()

	if cfg.PoolSize != 25 || cfg.MinIdleConns != 5 {
		t.Fatalf("pool sizes overwritten: PoolSize=%d MinIdleConns=%d", cfg.PoolSize, cfg.MinIdleConns)
	}
	if cfg.Port != 6380 {
		t.Fatalf("Port overwritten: %d", cfg.Port)
	}
	if cfg.ConnectTimeout != 10*time.Second {
		t.Fatalf("ConnectTimeout overwritten: %v", cfg.ConnectTimeout)
	}
}

func TestOptionsFromDiscreteFields(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		Host:     "cache.internal",
		Port:     6380,
		Username: "app",
		Password: "s3cret",
		DB:       2,
	}
	cfg.withDefaults()

	opts, err := cfg.options()
	if err != nil {
		t.Fatal(err)
	}

	if opts.Addr != "cache.internal:6380" {
		t.Fatalf("Addr = %q, want %q", opts.Addr, "cache.internal:6380")
	}
	if opts.Username != "app" || opts.Password != "s3cret" || opts.DB != 2 {
		t.Fatalf("auth/db wrong: user=%q pass=%q db=%d", opts.Username, opts.Password, opts.DB)
	}
	if opts.PoolSize != defaultPoolSize {
		t.Fatalf("PoolSize = %d, want %d", opts.PoolSize, defaultPoolSize)
	}
}

func TestOptionsPrefersDSN(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		DSN:      "redis://other:pass@localhost:6379/3",
		Host:     "ignored",
		Password: "ignored",
		DB:       0,
	}
	cfg.withDefaults()

	opts, err := cfg.options()
	if err != nil {
		t.Fatal(err)
	}
	if opts.Addr != "localhost:6379" {
		t.Fatalf("Addr = %q, want %q", opts.Addr, "localhost:6379")
	}
	if opts.Username != "other" || opts.Password != "pass" || opts.DB != 3 {
		t.Fatalf("DSN fields not applied: user=%q pass=%q db=%d", opts.Username, opts.Password, opts.DB)
	}
}
