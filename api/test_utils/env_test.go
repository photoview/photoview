package test_utils

import (
	"strings"
	"testing"
)

func TestPathFromAPIRoot(t *testing.T) {
	tests := []struct {
		paths []string
		want  string
	}{
		{[]string{"server.go"}, "/api/server.go"},
		{[]string{"scanner", "..", "server.go"}, "/api/server.go"},
		{[]string{"scanner", "scanner_test.go"}, "/api/scanner/scanner_test.go"},
	}

	for _, tc := range tests {
		if got, want := PathFromAPIRoot(tc.paths...), tc.want; !strings.HasSuffix(got, want) {
			t.Fatalf("PathFromAPIRoot(%v) = %q, want a suffix: %q", tc.paths, got, want)
		}
	}
}
