package postgres

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
			cfg:  &Config{DSN: "postgres://u:p@localhost:5432/db"},
		},
		{
			name: "discrete fields",
			cfg: &Config{
				Host:     "localhost",
				User:     "u",
				Database: "db",
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
		{
			name:    "host without user",
			cfg:     &Config{Host: "localhost", Database: "db"},
			wantErr: "User is required",
		},
		{
			name:    "host without database",
			cfg:     &Config{Host: "localhost", User: "u"},
			wantErr: "Database is required",
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

	cfg := &Config{DSN: "postgres://u:p@localhost/db"}
	cfg.withDefaults()

	if cfg.Port != defaultPort {
		t.Errorf("Port = %d, want %d", cfg.Port, defaultPort)
	}
	if cfg.SSLMode != defaultSSLMode {
		t.Errorf("SSLMode = %q, want %q", cfg.SSLMode, defaultSSLMode)
	}
	if cfg.MaxConns != defaultMaxConns {
		t.Errorf("MaxConns = %d, want %d", cfg.MaxConns, defaultMaxConns)
	}
	if cfg.MinConns != defaultMinConns {
		t.Errorf("MinConns = %d, want %d", cfg.MinConns, defaultMinConns)
	}
	if cfg.MaxConnLifetime != defaultMaxConnLifetime {
		t.Errorf("MaxConnLifetime = %v, want %v", cfg.MaxConnLifetime, defaultMaxConnLifetime)
	}
	if cfg.MaxConnIdleTime != defaultMaxConnIdleTime {
		t.Errorf("MaxConnIdleTime = %v, want %v", cfg.MaxConnIdleTime, defaultMaxConnIdleTime)
	}
	if cfg.HealthCheckPeriod != defaultHealthCheckPeriod {
		t.Errorf("HealthCheckPeriod = %v, want %v", cfg.HealthCheckPeriod, defaultHealthCheckPeriod)
	}
	if cfg.ConnectTimeout != defaultConnectTimeout {
		t.Errorf("ConnectTimeout = %v, want %v", cfg.ConnectTimeout, defaultConnectTimeout)
	}
}

func TestConfigWithDefaultsPreservesOverrides(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		DSN:               "postgres://u:p@localhost/db",
		MaxConns:          25,
		MinConns:          5,
		MaxConnLifetime:   2 * time.Hour,
		MaxConnIdleTime:   10 * time.Minute,
		HealthCheckPeriod: 15 * time.Second,
		ConnectTimeout:    10 * time.Second,
		SSLMode:           "require",
		Port:              6543,
	}
	cfg.withDefaults()

	if cfg.MaxConns != 25 || cfg.MinConns != 5 {
		t.Fatalf("pool sizes overwritten: MaxConns=%d MinConns=%d", cfg.MaxConns, cfg.MinConns)
	}
	if cfg.SSLMode != "require" || cfg.Port != 6543 {
		t.Fatalf("connection fields overwritten: SSLMode=%q Port=%d", cfg.SSLMode, cfg.Port)
	}
	if cfg.ConnectTimeout != 10*time.Second {
		t.Fatalf("ConnectTimeout overwritten: %v", cfg.ConnectTimeout)
	}
}

func TestConnStringFromDiscreteFields(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		Host:     "db.internal",
		Port:     5433,
		User:     "app",
		Password: "s3cret",
		Database: "ildoo",
		SSLMode:  "require",
	}
	cfg.withDefaults()

	got, err := cfg.connString()
	if err != nil {
		t.Fatal(err)
	}

	want := "postgres://app:s3cret@db.internal:5433/ildoo?sslmode=require"
	if got != want {
		t.Fatalf("connString = %q, want %q", got, want)
	}
}

func TestConnStringPrefersDSN(t *testing.T) {
	t.Parallel()

	dsn := "postgres://other@localhost/otherdb?sslmode=disable"
	cfg := &Config{
		DSN:      dsn,
		Host:     "ignored",
		User:     "ignored",
		Database: "ignored",
	}

	got, err := cfg.connString()
	if err != nil {
		t.Fatal(err)
	}
	if got != dsn {
		t.Fatalf("connString = %q, want %q", got, dsn)
	}
}
