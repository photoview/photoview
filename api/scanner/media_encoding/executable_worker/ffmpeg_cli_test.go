package executable_worker

import (
	"errors"
	"regexp"
	"testing"

	"github.com/photoview/photoview/api/utils"
	"github.com/spf13/afero"
	"gopkg.in/vansante/go-ffprobe.v2"
)

func TestFfmpegNotExist(t *testing.T) {
	fs := afero.NewOsFs()
	SetPathWithCurrent(t, "")

	Ffmpeg = newFfmpegCli()

	if got, want := Ffmpeg.err, ErrNoDependency; got != want {
		t.Errorf("Ffmpeg.err = %v, want: %v", got, want)
	}

	if Ffmpeg.IsInstalled() {
		t.Error("Ffmpeg should not be installed, but is found:", Ffmpeg)
	}

	if got, want := Ffmpeg.EncodeMp4(fs, "input", "output"), ErrNoDependency; !errors.Is(got, want) {
		t.Errorf("Ffmpge.EncodeMp4() = %v, want: %v", got, want)
	}

	if got, want := Ffmpeg.EncodeVideoThumbnail(fs, "input", "output", nil), ErrNoDependency; !errors.Is(got, want) {
		t.Errorf("Ffmpge.EncodeMp4() = %v, want: %v", got, want)
	}
}

func TestFfmpegVersionFail(t *testing.T) {
	fs := afero.NewOsFs()
	SetPathWithCurrent(t, testdataBinPath)
	t.Setenv("FAIL_WITH", "expect failure")

	Ffmpeg = newFfmpegCli()

	if got, want := Ffmpeg.err, ErrNoDependency; got != want {
		t.Errorf("Ffmpeg.err = %v, want: %v", got, want)
	}

	if Ffmpeg.IsInstalled() {
		t.Error("Ffmpeg should not be installed, but is found:", Ffmpeg)
	}

	if got, want := Ffmpeg.EncodeMp4(fs, "input", "output"), ErrNoDependency; !errors.Is(got, want) {
		t.Errorf("Ffmpge.EncodeMp4() = %v, want: %v", got, want)
	}

	if got, want := Ffmpeg.EncodeVideoThumbnail(fs, "input", "output", nil), ErrNoDependency; !errors.Is(got, want) {
		t.Errorf("Ffmpge.EncodeMp4() = %v, want: %v", got, want)
	}
}

func TestFfmpegIgnore(t *testing.T) {
	fs := afero.NewOsFs()
	SetPathWithCurrent(t, testdataBinPath)
	t.Setenv("PHOTOVIEW_DISABLE_VIDEO_ENCODING", "true")

	Ffmpeg = newFfmpegCli()

	if got, want := Ffmpeg.err, ErrDisabledFunction; got != want {
		t.Errorf("Ffmpeg.err = %v, want: %v", got, want)
	}

	if Ffmpeg.IsInstalled() {
		t.Error("Ffmpeg should be ignored (as it is disabled), but is initialized:", Ffmpeg)
	}

	if got, want := Ffmpeg.EncodeMp4(fs, "input", "output"), ErrDisabledFunction; !errors.Is(got, want) {
		t.Errorf("Ffmpge.EncodeMp4() = %v, want: %v", got, want)
	}

	if got, want := Ffmpeg.EncodeVideoThumbnail(fs, "input", "output", nil), ErrDisabledFunction; !errors.Is(got, want) {
		t.Errorf("Ffmpge.EncodeMp4() = %v, want: %v", got, want)
	}
}

func TestFfmpeg(t *testing.T) {
	fs := afero.NewOsFs()
	SetPathWithCurrent(t, testdataBinPath)

	Ffmpeg = newFfmpegCli()

	if !Ffmpeg.IsInstalled() {
		t.Fatal("Ffmpeg should be installed")
	}

	t.Run("EncodeMp4Failed", func(t *testing.T) {
		t.Setenv("FAIL_WITH", "expect failure")

		err := Ffmpeg.EncodeMp4(fs, "input", "output")
		if err == nil {
			t.Fatalf("Ffmpeg.EncodeMp4(...) = nil, should be an error.")
		}
		if got, want := err.Error(), `^encoding video with ".*/test_data/mock_bin/ffmpeg" \[-i input -vcodec h264 .* output\] error: .*$`; !regexp.MustCompile(want).MatchString(got) {
			t.Errorf("Ffmpeg.EncodeMp4(...) = %q, should be as reg pattern %q", got, want)
		}
	})

	t.Run("EncodeMp4Succeeded", func(t *testing.T) {
		err := Ffmpeg.EncodeMp4(fs, "input", "output")
		if err != nil {
			t.Fatalf("Ffmpeg.EncodeMp4(...) = %v, should be nil.", err)
		}
	})

	probeData := &ffprobe.ProbeData{
		Format: &ffprobe.Format{
			DurationSeconds: 10,
		},
	}
	t.Run("EncodeVideoThumbnailMp4Failed", func(t *testing.T) {
		t.Setenv("FAIL_WITH", "expect failure")

		err := Ffmpeg.EncodeVideoThumbnail(fs, "input", "output", probeData)
		if err == nil {
			t.Fatalf("Ffmpeg.EncodeVideoThumbnail(...) = nil, should be an error.")
		}
		if got, want := err.Error(), `^encoding video thumbnail with ".*/test_data/mock_bin/ffmpeg" \[-ss 2 -i input .* output\] error: .*$`; !regexp.MustCompile(want).MatchString(got) {
			t.Errorf("Ffmpeg.EncodeVideoThumbnail(...) = %q, should be as reg pattern %q", got, want)
		}
	})

	t.Run("EncodeVideoThumbnailSucceeded", func(t *testing.T) {
		err := Ffmpeg.EncodeVideoThumbnail(fs, "input", "output", probeData)
		if err != nil {
			t.Fatalf("Ffmpeg.EncodeVideoThumbnail(...) = %v, should be nil.", err)
		}
	})
}

func TestFfmpegWithHWAcc(t *testing.T) {
	fs := afero.NewOsFs()
	SetPathWithCurrent(t, testdataBinPath)
	t.Setenv(utils.EnvVideoHardwareAcceleration.GetName(), "qsv")

	Ffmpeg = newFfmpegCli()

	t.Setenv("FAIL_WITH", "expect failure")

	err := Ffmpeg.EncodeMp4(fs, "input", "output")
	if err == nil {
		t.Fatalf("Ffmpeg.EncodeMp4(...) = nil, should be an error.")
	}
	if got, want := err.Error(), `^encoding video with ".*/test_data/mock_bin/ffmpeg" \[-i input -vcodec h264_qsv .* output\] error: .*$`; !regexp.MustCompile(want).MatchString(got) {
		t.Errorf("Ffmpeg.EncodeMp4(...) = %q, should be as reg pattern %q", got, want)
	}
}

func TestFfmpegWithCustomCodec(t *testing.T) {
	fs := afero.NewOsFs()
	SetPathWithCurrent(t, testdataBinPath)
	t.Setenv(utils.EnvVideoHardwareAcceleration.GetName(), "_custom")

	Ffmpeg = newFfmpegCli()

	t.Setenv("FAIL_WITH", "expect failure")

	err := Ffmpeg.EncodeMp4(fs, "input", "output")
	if err == nil {
		t.Fatalf("Ffmpeg.EncodeMp4(...) = nil, should be an error.")
	}
	if got, want := err.Error(), `^encoding video with ".*/test_data/mock_bin/ffmpeg" \[-i input -vcodec custom .* output\] error: .*$`; !regexp.MustCompile(want).MatchString(got) {
		t.Errorf("Ffmpeg.EncodeMp4(...) = %q, should be as reg pattern %q", got, want)
	}
}
