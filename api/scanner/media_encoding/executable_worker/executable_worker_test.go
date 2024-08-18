package executable_worker_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/photoview/photoview/api/test_utils"
)

func TestMain(m *testing.M) {
	os.Exit(test_utils.IntegrationTestRun(m))
}

func setPathWithCurrent(paths ...string) func() {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return func() {}
	}

	base := filepath.Dir(file)

	for i, path := range paths {
		paths[i] = filepath.Join(base, path)
	}

	originalPath := os.Getenv("PATH")
	os.Setenv("PATH", strings.Join(paths, ":"))

	return func() {
		os.Setenv("PATH", originalPath)
	}
}
