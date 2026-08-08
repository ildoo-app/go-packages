package logger

import (
	"context"
	"log/slog"
)

type ctxKey struct{}

// WithContext returns a child context carrying log.
func WithContext(ctx context.Context, log Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, log)
}

// FromContext returns the logger stored in ctx, or a wrapper around slog.Default
// if none is set.
func FromContext(ctx context.Context) Logger {
	if ctx != nil {
		if log, ok := ctx.Value(ctxKey{}).(Logger); ok && log != nil {
			return log
		}
	}
	return &slogLogger{Logger: slog.Default()}
}
