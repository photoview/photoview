package log

import (
	"context"
)

// Debug logs debug messages.
func Debug(ctx context.Context, msg string, args ...any) {
	getLogger(ctx).DebugContext(ctx, msg, args...)
}

// Info logs info messages.
func Info(ctx context.Context, msg string, args ...any) {
	getLogger(ctx).InfoContext(ctx, msg, args...)
}

// Warn logs warning messages.
func Warn(ctx context.Context, msg string, args ...any) {
	getLogger(ctx).WarnContext(ctx, msg, args...)
}

// Error logs error messages.
func Error(ctx context.Context, msg string, args ...any) {
	getLogger(ctx).ErrorContext(ctx, msg, args...)
}
