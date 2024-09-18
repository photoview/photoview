package executable_worker_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/photoview/photoview/api/scanner/media_encoding/executable_worker"
	"github.com/photoview/photoview/api/test_utils"
)

func TestMain(m *testing.M) {
	os.Exit(test_utils.IntegrationTestRun(m))
}

func setPathWithCurrent(paths ...string) func() {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return func() {
			// Return an empty function in case of error
		}
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

func setEnv(key, value string) func() {
	org := os.Getenv(key)
	os.Setenv(key, value)
	return func() {
		os.Setenv(key, org)
	}
}

func TestInitFfprobePath(t *testing.T) {
	t.Run("PathFail", func(t *testing.T) {
		err := executable_worker.InitFfprobePath()
		if err == nil {
			t.Fatalf("InitFfprobePath() returns nil, want an error")
		}
	})

	t.Run("VersionFail", func(t *testing.T) {
		donePath := setPathWithCurrent("./testdata/bin")
		defer donePath()

		doneEnv := setEnv("FAIL_WITH", "expect failure")
		defer doneEnv()

		err := executable_worker.InitFfprobePath()
		if err == nil {
			t.Fatalf("InitFfprobePath() returns nil, want an error")
		}
	})

	t.Run("Succeed", func(t *testing.T) {
		donePath := setPathWithCurrent("./testdata/bin")
		defer donePath()

		err := executable_worker.InitFfprobePath()
		if err != nil {
			t.Fatalf("InitFfprobePath() returns %v, want nil", err)
		}
	})
}
