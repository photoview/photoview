package executable_worker_test

import (
	"regexp"
	"testing"

	"github.com/photoview/photoview/api/scanner/media_encoding/executable_worker"
)

func TestMagickCliNotExist(t *testing.T) {
	done := setPathWithCurrent()
	defer done()

	executable_worker.InitializeExecutableWorkers()
	if executable_worker.Magick.IsInstalled() {
		t.Error("MagickCli should not be installed, but is found:", executable_worker.Magick)
	}
}

func TestMagickCliIgnore(t *testing.T) {
	donePath := setPathWithCurrent("./testdata/bin")
	defer donePath()

	doneDisableRaw := setEnv("PHOTOVIEW_DISABLE_RAW_PROCESSING", "true")
	defer doneDisableRaw()

	executable_worker.InitializeExecutableWorkers()
	if executable_worker.Magick.IsInstalled() {
		t.Error("MagickCli should not be installed, but is found:", executable_worker.Magick)
	}
}

func TestMagickCliFail(t *testing.T) {
	donePath := setPathWithCurrent("./testdata/bin")
	defer donePath()

	executable_worker.InitializeExecutableWorkers()
	if !executable_worker.Magick.IsInstalled() {
		t.Fatal("MagickCli should be installed")
	}

	done := setEnv("FAIL_WITH", "failure")
	defer done()

	err := executable_worker.Magick.EncodeJpeg("input", "output", 70)
	if err == nil {
		t.Fatalf(`MagickCli.EncodeJpeg(...) = nil, should be an error.`)
	}

	if got, want := err.Error(), `^encoding image with ".*/testdata/bin/magick \[convert input -quality 70 output\]" error: .*$`; !regexp.MustCompile(want).MatchString(got) {
		t.Errorf(`MagickCli.EncodeJpeg(...) = %q, should be matched with reg pattern %q`, got, want)
	}
}

func TestMagickCliSucceed(t *testing.T) {
	donePath := setPathWithCurrent("./testdata/bin")
	defer donePath()

	executable_worker.InitializeExecutableWorkers()
	if !executable_worker.Magick.IsInstalled() {
		t.Fatal("MagickCli should be installed")
	}

	t.Run("Succeeded", func(t *testing.T) {
		err := executable_worker.Magick.EncodeJpeg("input", "output", 70)
		if err != nil {
			t.Fatalf("MagickCli.EncodeJpeg(...) = %v, should be nil.", err)
		}
	})
}
