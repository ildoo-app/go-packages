package lifecycle

import (
	"errors"
	"time"
)

const defaultGraceTimeout = 15 * time.Second

// Config holds process lifecycle settings.
// Callers map their own env vars into this struct.
type Config struct {
	// GraceTimeout bounds the whole teardown, including every DI cleanup.
	// Must be shorter than the container stop grace period. Defaults to 15s.
	GraceTimeout time.Duration
}

func (c *Config) validate() error {
	if c == nil {
		return errors.New("lifecycle: config is nil")
	}
	if c.GraceTimeout < 0 {
		return errors.New("lifecycle: GraceTimeout must be non-negative")
	}
	return nil
}

func (c *Config) withDefaults() {
	if c.GraceTimeout == 0 {
		c.GraceTimeout = defaultGraceTimeout
	}
}
