package lifecycle

import (
	"strings"
	"testing"
	"time"
)

func TestConfigValidateRejectsNil(t *testing.T) {
	t.Parallel()

	var cfg *Config
	if err := cfg.validate(); err == nil || !strings.Contains(err.Error(), "config is nil") {
		t.Fatalf("got %v, want config is nil", err)
	}
}

func TestConfigValidateRejectsNegativeGrace(t *testing.T) {
	t.Parallel()

	cfg := &Config{GraceTimeout: -time.Second}
	if err := cfg.validate(); err == nil || !strings.Contains(err.Error(), "non-negative") {
		t.Fatalf("got %v, want non-negative", err)
	}
}

func TestConfigWithDefaults(t *testing.T) {
	t.Parallel()

	cfg := &Config{}
	cfg.withDefaults()
	if cfg.GraceTimeout != defaultGraceTimeout {
		t.Fatalf("got %v, want %v", cfg.GraceTimeout, defaultGraceTimeout)
	}
}
