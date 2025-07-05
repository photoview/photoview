package log

import (
	"bytes"
	"context"
	"flag"
	"log/slog"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func init() {
	// Avoid panic with providing flags in `test_utils/integration_setup.go`.
	flag.CommandLine.Init("log", flag.ContinueOnError)
}

func mockDefaultLogger(t *testing.T) *bytes.Buffer {
	t.Helper()

	var output bytes.Buffer
	handler := slog.NewJSONHandler(&output, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Remove time from the output for predictable test output.
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	})

	logger := slog.New(handler)
	oldLogger := defaultLogger
	defaultLogger = logger

	t.Cleanup(func() {
		defaultLogger = oldLogger
	})

	return &output
}

func TestLogger(t *testing.T) {
	output := mockDefaultLogger(t)

	Debug(nil, "no_context")
	Info(nil, "no_context")
	Warn(nil, "no_context")
	Error(nil, "no_context")

	bg := context.Background()

	Debug(bg, "bg_context")
	Info(bg, "bg_context")
	Warn(bg, "bg_context")
	Error(bg, "bg_context")

	ctx1 := WithAttrs(context.Background(), "arg1", "value")
	Debug(ctx1, "with_context_1")
	Info(ctx1, "with_context_1")
	Warn(ctx1, "with_context_1")
	Error(ctx1, "with_context_1")

	ctx2 := WithAttrs(ctx1, "arg2", "value")
	Debug(ctx2, "with_context_2")
	Info(ctx2, "with_context_2")
	Warn(ctx2, "with_context_2")
	Error(ctx2, "with_context_2")

	Debug(ctx1, "with_context_1")
	Info(ctx1, "with_context_1")
	Warn(ctx1, "with_context_1")
	Error(ctx1, "with_context_1")

	// Debug messages are filtered out due to slog.LevelInfo in mock logger
	want := `{"level":"INFO","msg":"no_context"}
{"level":"WARN","msg":"no_context"}
{"level":"ERROR","msg":"no_context"}
{"level":"INFO","msg":"bg_context"}
{"level":"WARN","msg":"bg_context"}
{"level":"ERROR","msg":"bg_context"}
{"level":"INFO","msg":"with_context_1","arg1":"value"}
{"level":"WARN","msg":"with_context_1","arg1":"value"}
{"level":"ERROR","msg":"with_context_1","arg1":"value"}
{"level":"INFO","msg":"with_context_2","arg1":"value","arg2":"value"}
{"level":"WARN","msg":"with_context_2","arg1":"value","arg2":"value"}
{"level":"ERROR","msg":"with_context_2","arg1":"value","arg2":"value"}
{"level":"INFO","msg":"with_context_1","arg1":"value"}
{"level":"WARN","msg":"with_context_1","arg1":"value"}
{"level":"ERROR","msg":"with_context_1","arg1":"value"}
`
	if diff := cmp.Diff(output.String(), want); diff != "" {
		t.Errorf("diff: (-got, +want)\n%s", diff)
	}

}
