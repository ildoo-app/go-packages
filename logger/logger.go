package logger

import (
	"log/slog"
	"os"
)

// slogLogger wraps *slog.Logger so With/WithGroup return Logger.
type slogLogger struct {
	*slog.Logger
}

func (l *slogLogger) With(args ...any) Logger {
	return &slogLogger{Logger: l.Logger.With(args...)}
}

func (l *slogLogger) WithGroup(name string) Logger {
	return &slogLogger{Logger: l.Logger.WithGroup(name)}
}

// New builds a Logger from cfg.
func New(cfg *Config) (Logger, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	cfg.withDefaults()

	w := cfg.Writer
	if w == nil {
		w = os.Stdout
	}

	opts := &slog.HandlerOptions{
		Level:     cfg.Level,
		AddSource: cfg.AddSource,
	}

	var handler slog.Handler
	switch cfg.Format {
	case FormatText:
		handler = slog.NewTextHandler(w, opts)
	default:
		handler = slog.NewJSONHandler(w, opts)
	}

	log := slog.New(handler)
	if cfg.Service != "" {
		log = log.With(slog.String("service", cfg.Service))
	}
	return &slogLogger{Logger: log}, nil
}
