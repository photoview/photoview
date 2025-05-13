package executable_worker

import (
	"errors"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/photoview/photoview/api/scanner/media_encoding/media_utils"
	"github.com/photoview/photoview/api/test_utils"
)

func TestMagickCliNotExist(t *testing.T) {
	done := test_utils.SetPathWithCurrent()
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

	if got, want := Magick.GenerateThumbnail("input", "output", 100, 100), ErrNoDependency; !errors.Is(got, want) {
		t.Errorf("Magick.GenerateThumbnail() = %v, want: %v", got, want)
	}

	{
		_, got := Magick.IdentifyDimension("input")
		if want := ErrNoDependency; !errors.Is(got, want) {
			t.Errorf("Magick.IdentifyDimension() = %v, want: %v", got, want)
		}
	}
}

func TestMagickCliIgnore(t *testing.T) {
	donePath := test_utils.SetPathWithCurrent(testdataBinPath)
	defer donePath()

	doneDisableRaw := test_utils.SetEnv("PHOTOVIEW_DISABLE_RAW_PROCESSING", "true")
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

	if got, want := Magick.GenerateThumbnail("input", "output", 100, 100), ErrDisabledFunction; !errors.Is(got, want) {
		t.Errorf("Magick.GenerateThumbnail() = %v, want: %v", got, want)
	}

	{
		_, got := Magick.IdentifyDimension("input")
		if want := ErrDisabledFunction; !errors.Is(got, want) {
			t.Errorf("Magick.IdentifyDimension() = %v, want: %v", got, want)
		}
	}
}

func TestMagickCliVersionFail(t *testing.T) {
	donePath := test_utils.SetPathWithCurrent(testdataBinPath)
	defer donePath()

	done := test_utils.SetEnv("FAIL_WITH", "failure")
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

	if got, want := Magick.GenerateThumbnail("input", "output", 100, 100), ErrNoDependency; !errors.Is(got, want) {
		t.Errorf("Magick.GenerateThumbnail() = %v, want: %v", got, want)
	}

	{
		_, got := Magick.IdentifyDimension("input")
		if want := ErrNoDependency; !errors.Is(got, want) {
			t.Errorf("Magick.IdentifyDimension() = %v, want: %v", got, want)
		}
	}
}

func TestMagickCliFail(t *testing.T) {
	donePath := test_utils.SetPathWithCurrent(testdataBinPath)
	defer donePath()

	Magick = newMagickCli()

	if !Magick.IsInstalled() {
		t.Fatal("MagickCli should be installed")
	}

	done := test_utils.SetEnv("FAIL_WITH", "failure")
	defer done()

	err := Magick.EncodeJpeg("input", "output", 70)
	if err == nil {
		t.Fatalf(`MagickCli.EncodeJpeg(...) = nil, should be an error.`)
	}

	if got, want := err.Error(), `^encoding image with ".*/test_data/mock_bin/magick \[input -auto-orient -quality 70 output\]" error: .*$`; !regexp.MustCompile(want).MatchString(got) {
		t.Errorf(`MagickCli.EncodeJpeg(...) = %q, should be matched with reg pattern %q`, got, want)
	}

	err = Magick.GenerateThumbnail("input", "output", 100, 100)
	if err == nil {
		t.Fatalf(`MagickCli.GenerateThumbnail(...) = nil, should be an error.`)
	}
	if got, want := err.Error(), `^generate thumbnail with ".*/test_data/mock_bin/magick \[input -resize 100x100 output\]" error: .*$`; !regexp.MustCompile(want).MatchString(got) {
		t.Errorf(`MagickCli.GenerateThumbnail(...) = %q, should be matched with reg pattern %q`, got, want)
	}

	{
		_, got := Magick.IdentifyDimension("input")
		if want := `^identify dimension with ".*/test_data/mock_bin/magick \[identify -format {"height":\%h, "width":\%w} input\]" error: .*$`; !regexp.MustCompile(want).MatchString(got.Error()) {
			t.Errorf("Magick.IdentifyDimension() = %v, should be matched with reg pattern %q", got, want)
		}
	}
}

func TestMagickCliSucceed(t *testing.T) {
	donePath := test_utils.SetPathWithCurrent(testdataBinPath)
	defer donePath()

	Magick = newMagickCli()

	if !Magick.IsInstalled() {
		t.Fatal("MagickCli should be installed")
	}

	t.Run("EncodeJpeg", func(t *testing.T) {
		err := Magick.EncodeJpeg("input", "output", 70)
		if err != nil {
			t.Fatalf("MagickCli.EncodeJpeg(...) = %v, should be nil.", err)
		}
	})

	t.Run("GenerateThumbnail", func(t *testing.T) {
		err := Magick.GenerateThumbnail("input", "output", 100, 100)
		if err != nil {
			t.Fatalf("MagickCli.GenerateThumbnail(...) = %v, should be nil.", err)
		}
	})

	t.Run("IdentifyDimension", func(t *testing.T) {
		got, err := Magick.IdentifyDimension("input")
		if err != nil {
			t.Fatalf("MagickCli.IdentifyDimension(...) = %v, should be nil.", err)
		}

		if diff := cmp.Diff(got, media_utils.PhotoDimensions{Width: 1000, Height: 800}); diff != "" {
			t.Errorf("MagickCli.IdentifyDimension(...) diff: (-got, +want)\n%s", diff)
		}
	})
}
