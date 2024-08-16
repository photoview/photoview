package executable_worker_test

import (
	"os"
	"path/filepath"
	"regexp"
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

func TestMagickWorkerNotExist(t *testing.T) {
	done := setPathWithCurrent()
	defer done()

	executable_worker.InitializeExecutableWorkers()

	if executable_worker.MagickCli.IsInstalled() {
		t.Error("MagickCli should not be installed, but is found:", executable_worker.MagickCli)
	}
}

func TestMagickWorkerIgnore(t *testing.T) {
	done := setPathWithCurrent("./testdata/bin")
	defer done()

	org := os.Getenv("PHOTOVIEW_DISABLE_RAW_PROCESSING")
	os.Setenv("PHOTOVIEW_DISABLE_RAW_PROCESSING", "true")
	defer os.Setenv("PHOTOVIEW_DISABLE_RAW_PROCESSING", org)

	executable_worker.InitializeExecutableWorkers()

	if executable_worker.MagickCli.IsInstalled() {
		t.Error("MagickCli should not be installed, but is found:", executable_worker.MagickCli)
	}
}

func TestMagickWorker(t *testing.T) {
	done := setPathWithCurrent("./testdata/bin")
	defer done()

	executable_worker.InitializeExecutableWorkers()

	if !executable_worker.MagickCli.IsInstalled() {
		t.Error("MagickCli should be installed")
	}

	t.Run("Failed", func(t *testing.T) {
		err := executable_worker.MagickCli.EncodeJpeg("input", "output", 0)
		if err == nil {
			t.Fatalf("MagickCli.EncodeJpeg(\"input\", \"output\", 0) = nil, should be an error.")
		}
		if got, want := err.Error(), "^encoding image with \".*?/testdata/bin/convert .*?\" error: .*$"; !regexp.MustCompile(want).MatchString(got) {
			t.Errorf("MagickCli.EncodeJpeg(\"input\", \"output\", 0) = %q, should be as reg pattern %q", got, want)
		}
	})

	t.Run("Succeeded", func(t *testing.T) {
		err := executable_worker.MagickCli.EncodeJpeg("input", "output", 70)
		if err != nil {
			t.Fatalf("MagickCli.EncodeJpeg(\"input\", \"output\", 0) = %v, should be nil.", err)
		}
	})
}
