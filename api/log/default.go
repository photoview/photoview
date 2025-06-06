package log

import (
	"context"
)

// Debug calls [Logger.DebugContext] on the default logger.
func Debug(ctx context.Context, msg string, args ...any) {
	getLogger(ctx).DebugContext(ctx, msg, args...)
}

// Info calls [Logger.InfoContext] on the default logger.
func Info(ctx context.Context, msg string, args ...any) {
	getLogger(ctx).InfoContext(ctx, msg, args...)
}

// Warn calls [Logger.WarnContext] on the default logger.
func Warn(ctx context.Context, msg string, args ...any) {
	getLogger(ctx).WarnContext(ctx, msg, args...)
}

// Error calls [Logger.ErrorContext] on the default logger.
func Error(ctx context.Context, msg string, args ...any) {
	getLogger(ctx).ErrorContext(ctx, msg, args...)
}
