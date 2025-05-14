package test_utils

import (
	"strings"
	"testing"
)

func TestPathFromAPIRoot(t *testing.T) {
	if got, want := PathFromAPIRoot("./server.go"), "/api/server.go"; !strings.HasSuffix(got, want) {
		t.Fatalf(`PathFromAPIRoot("./server.go") = %q, want a suffix: %q`, got, want)
	}
}
