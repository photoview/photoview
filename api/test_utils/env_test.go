package test_utils

import (
	"os"
	"strings"
	"testing"
)

func TestPathFromAPIRoot(t *testing.T) {
	if got, want := PathFromAPIRoot("./server.go"), "/api/server.go"; !strings.HasSuffix(got, want) {
		t.Fatalf(`PathFromAPIRoot("./server.go") = %q, want a suffix: %q`, got, want)
	}
}

func TestSetPathWithCurrent(t *testing.T) {
	pathDone := SetEnv("PATH", "")
	defer pathDone()

	testDone := SetPathWithCurrent("./test")
	defer testDone()

	path := os.Getenv("PATH")
	if got, want := path, "api/test_utils/test"; !strings.HasSuffix(got, want) {
		t.Errorf("path = %q, want a suffix: %q", got, want)
	}
}

func TestAlwaysFail(t *testing.T) {
	t.Fatal("always fail")
}
