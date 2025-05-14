package executable_worker

import (
	"errors"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMagickCliNotExist(t *testing.T) {
	SetPathWithCurrent(t, "")

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
	SetPathWithCurrent(t, testdataBinPath)
	t.Setenv("PHOTOVIEW_DISABLE_RAW_PROCESSING", "true")

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
	SetPathWithCurrent(t, testdataBinPath)
	t.Setenv("FAIL_WITH", "failure")

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
	SetPathWithCurrent(t, testdataBinPath)

	Magick = newMagickCli()

	if !Magick.IsInstalled() {
		t.Fatal("MagickCli should be installed")
	}

	t.Setenv("FAIL_WITH", "failure")

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
	SetPathWithCurrent(t, testdataBinPath)

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

		if diff := cmp.Diff(got, Dimension{Width: 1000, Height: 800}); diff != "" {
			t.Errorf("MagickCli.IdentifyDimension(...) diff: (-got, +want)\n%s", diff)
		}
	})

	t.Run("IdentifyDimensionInvalidJSON", func(t *testing.T) {
		t.Setenv("INVALID_OUTPUT", `{"width":1000,`)

		_, err := Magick.IdentifyDimension("input")
		if want := `unexpected EOF$`; !regexp.MustCompile(want).MatchString(err.Error()) {
			t.Errorf("MagickCli.IdentifyDimension() = error(%v), which should match with regexp %q", err, want)
		}
	})
}
