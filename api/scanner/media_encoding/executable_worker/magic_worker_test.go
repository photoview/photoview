package executable_worker_test

import (
	"os"
	"regexp"
	"testing"

	"github.com/photoview/photoview/api/scanner/media_encoding/executable_worker"
)

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

	envKey := "PHOTOVIEW_DISABLE_RAW_PROCESSING"
	org := os.Getenv(envKey)
	os.Setenv(envKey, "true")
	defer os.Setenv(envKey, org)

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
		if got, want := err.Error(), `^encoding image with ".*?/testdata/bin/convert" .*? error: .*$`; !regexp.MustCompile(want).MatchString(got) {
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
