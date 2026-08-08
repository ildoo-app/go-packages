package logger

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
)

// Output format for the handler.
const (
	FormatJSON = "json"
	FormatText = "text"
)

// Config holds structured logger settings.
// Callers map their own env vars into this struct.
type Config struct {
	// Level is the minimum level that will be logged.
	// Defaults to slog.LevelInfo.
	Level slog.Level

	// Format is "json" (default) or "text".
	Format string

	// AddSource attaches source file/line to each record.
	AddSource bool

	// Writer receives log output. Defaults to os.Stdout.
	Writer io.Writer

	// Service is attached as a default "service" attribute when non-empty.
	Service string
}

func (c *Config) validate() error {
	if c == nil {
		return errors.New("logger: config is nil")
	}
	switch strings.ToLower(c.Format) {
	case "", FormatJSON, FormatText:
		return nil
	default:
		return fmt.Errorf("logger: unsupported format %q (want %q or %q)", c.Format, FormatJSON, FormatText)
	}
}

func (c *Config) withDefaults() {
	if c.Format == "" {
		c.Format = FormatJSON
	} else {
		c.Format = strings.ToLower(c.Format)
	}
	// Level zero value is already slog.LevelInfo (0).
}

// ParseLevel converts common level names to slog.Level.
// Accepts: debug, info, warn/warning, error (case-insensitive).
func ParseLevel(s string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info", "":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("logger: unknown level %q", s)
	}
}
