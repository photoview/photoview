package executable_worker

import (
	"errors"
	"regexp"
	"testing"

	"github.com/photoview/photoview/api/test_utils/test_env"
)

func TestMagickCliNotExist(t *testing.T) {
	done := test_env.SetPathWithCurrent()
	defer done()

	Magick = newMagickCli()

	if got, want := Magick.err, ErrNoDependency; got != want {
		t.Errorf("Magick.err = %v, want: %v", got, want)
	}

	if Magick.IsInstalled() {
		t.Error("MagickCli should not be installed, but is found:", Magick)
	}

	if got, want := Magick.EncodeJpeg("input", "output", 70), ErrNoDependency; !errors.Is(got, want) {
		t.Errorf("Magick.EncodeJpeg() = %v, want: %v", got, want)
	}
}

func TestMagickCliIgnore(t *testing.T) {
	donePath := test_env.SetPathWithCurrent(testdataBinPath)
	defer donePath()

	doneDisableRaw := test_env.SetEnv("PHOTOVIEW_DISABLE_RAW_PROCESSING", "true")
	defer doneDisableRaw()

	Magick = newMagickCli()

	if got, want := Magick.err, ErrDisabledFunction; got != want {
		t.Errorf("Magick.err = %v, want: %v", got, want)
	}

	if Magick.IsInstalled() {
		t.Error("MagickCli should not be installed, but is found:", Magick)
	}

	if got, want := Magick.EncodeJpeg("input", "output", 70), ErrDisabledFunction; !errors.Is(got, want) {
		t.Errorf("Magick.EncodeJpeg() = %v, want: %v", got, want)
	}
}

func TestMagickCliVersionFail(t *testing.T) {
	donePath := test_env.SetPathWithCurrent(testdataBinPath)
	defer donePath()

	done := test_env.SetEnv("FAIL_WITH", "failure")
	defer done()

	Magick = newMagickCli()

	if got, want := Magick.err, ErrNoDependency; got != want {
		t.Errorf("Magick.err = %v, want: %v", got, want)
	}

	if Magick.IsInstalled() {
		t.Error("MagickCli should not be installed, but is found:", Magick)
	}

	if got, want := Magick.EncodeJpeg("input", "output", 70), ErrNoDependency; !errors.Is(got, want) {
		t.Errorf("Magick.EncodeJpeg() = %v, want: %v", got, want)
	}
}

func TestMagickCliFail(t *testing.T) {
	donePath := test_env.SetPathWithCurrent(testdataBinPath)
	defer donePath()

	Magick = newMagickCli()

	if !Magick.IsInstalled() {
		t.Fatal("MagickCli should be installed")
	}

	done := test_env.SetEnv("FAIL_WITH", "failure")
	defer done()

	err := Magick.EncodeJpeg("input", "output", 70)
	if err == nil {
		t.Fatalf(`MagickCli.EncodeJpeg(...) = nil, should be an error.`)
	}

	if got, want := err.Error(), `^encoding image with ".*/test_data/bin/magick \[input -auto-orient -quality 70 output\]" error: .*$`; !regexp.MustCompile(want).MatchString(got) {
		t.Errorf(`MagickCli.EncodeJpeg(...) = %q, should be matched with reg pattern %q`, got, want)
	}
}

func TestMagickCliSucceed(t *testing.T) {
	donePath := test_env.SetPathWithCurrent(testdataBinPath)
	defer donePath()

	Magick = newMagickCli()

	if !Magick.IsInstalled() {
		t.Fatal("MagickCli should be installed")
	}

	t.Run("Succeeded", func(t *testing.T) {
		err := Magick.EncodeJpeg("input", "output", 70)
		if err != nil {
			t.Fatalf("MagickCli.EncodeJpeg(...) = %v, should be nil.", err)
		}
	})
}
