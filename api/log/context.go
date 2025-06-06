package log

import (
	"context"
	"log/slog"
)

var defaultLogger *slog.Logger

func init() {
	defaultLogger = slog.Default()
}

type loggerKeyType string

const loggerKey loggerKeyType = "logger"

func getLogger(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return defaultLogger
	}

	logger := ctx.Value(loggerKey)
	if logger == nil {
		return defaultLogger
	}

	ret, ok := logger.(*slog.Logger)
	if !ok {
		return defaultLogger
	}

	return ret
}

func WithAttrs(ctx context.Context, args ...any) context.Context {
	old := getLogger(ctx)
	new := old.With(args...)
	return context.WithValue(ctx, loggerKey, new)
}
