package server

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
			name: "empty config ok",
			cfg:  &Config{},
		},
		{
			name: "explicit port",
			cfg:  &Config{Port: 9090},
		},
		{
			name:    "nil config",
			cfg:     nil,
			wantErr: "config is nil",
		},
		{
			name:    "negative port",
			cfg:     &Config{Port: -1},
			wantErr: "out of range",
		},
		{
			name:    "port too large",
			cfg:     &Config{Port: 65536},
			wantErr: "out of range",
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

	cfg := &Config{}
	cfg.withDefaults()

	if cfg.Port != defaultPort {
		t.Errorf("Port = %d, want %d", cfg.Port, defaultPort)
	}
	if cfg.ShutdownTimeout != defaultShutdownTimeout {
		t.Errorf("ShutdownTimeout = %v, want %v", cfg.ShutdownTimeout, defaultShutdownTimeout)
	}
	if cfg.ReadHeaderTimeout != defaultReadHeaderTimeout {
		t.Errorf("ReadHeaderTimeout = %v, want %v", cfg.ReadHeaderTimeout, defaultReadHeaderTimeout)
	}
	if cfg.IdleTimeout != defaultIdleTimeout {
		t.Errorf("IdleTimeout = %v, want %v", cfg.IdleTimeout, defaultIdleTimeout)
	}
}

func TestConfigWithDefaultsPreservesOverrides(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		Host:              "127.0.0.1",
		Port:              9090,
		ShutdownTimeout:   5 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadTimeout:       3 * time.Second,
		WriteTimeout:      4 * time.Second,
	}
	cfg.withDefaults()

	if cfg.Host != "127.0.0.1" || cfg.Port != 9090 {
		t.Fatalf("bind overwritten: Host=%q Port=%d", cfg.Host, cfg.Port)
	}
	if cfg.ShutdownTimeout != 5*time.Second {
		t.Fatalf("ShutdownTimeout overwritten: %v", cfg.ShutdownTimeout)
	}
	if cfg.ReadHeaderTimeout != 2*time.Second || cfg.IdleTimeout != 30*time.Second {
		t.Fatalf("timeouts overwritten: ReadHeaderTimeout=%v IdleTimeout=%v", cfg.ReadHeaderTimeout, cfg.IdleTimeout)
	}
	if cfg.ReadTimeout != 3*time.Second || cfg.WriteTimeout != 4*time.Second {
		t.Fatalf("optional timeouts overwritten: ReadTimeout=%v WriteTimeout=%v", cfg.ReadTimeout, cfg.WriteTimeout)
	}
}

func TestConfigAddr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cfg  Config
		want string
	}{
		{name: "all interfaces", cfg: Config{Port: 8080}, want: ":8080"},
		{name: "host and port", cfg: Config{Host: "127.0.0.1", Port: 9090}, want: "127.0.0.1:9090"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.cfg.addr(); got != tt.want {
				t.Fatalf("addr = %q, want %q", got, tt.want)
			}
		})
	}
}
