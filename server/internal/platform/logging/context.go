package logging

import (
	"context"
	"log/slog"
)

type loggerKey struct{}

// WithLogger stores a logger in the context for downstream retrieval.
// If logger is nil, the context is returned unchanged so that
// FromContext can fall back to the default logger.
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	if logger == nil {
		return ctx
	}
	return context.WithValue(ctx, loggerKey{}, logger)
}

// FromContext extracts a logger from the context.
// Falls back to the global default logger if none is stored.
func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}
