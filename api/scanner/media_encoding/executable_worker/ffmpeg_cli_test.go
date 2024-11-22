package executable_worker

import (
	"errors"
	"regexp"
	"testing"

	"github.com/photoview/photoview/api/test_utils/test_env"
	"github.com/photoview/photoview/api/utils"
	"gopkg.in/vansante/go-ffprobe.v2"
)

func TestFfmpegNotExist(t *testing.T) {
	done := test_env.SetPathWithCurrent()
	defer done()

	Ffmpeg = newFfmpegCli()

	if got, want := Ffmpeg.err, ErrNoDependency; got != want {
		t.Errorf("Ffmpeg.err = %v, want: %v", got, want)
	}

	if Ffmpeg.IsInstalled() {
		t.Error("Ffmpeg should not be installed, but is found:", Ffmpeg)
	}

	if got, want := Ffmpeg.EncodeMp4("input", "output"), ErrNoDependency; !errors.Is(got, want) {
		t.Errorf("Ffmpge.EncodeMp4() = %v, want: %v", got, want)
	}

	if got, want := Ffmpeg.EncodeVideoThumbnail("input", "output", nil), ErrNoDependency; !errors.Is(got, want) {
		t.Errorf("Ffmpge.EncodeMp4() = %v, want: %v", got, want)
	}
}

func TestFfmpegVersionFail(t *testing.T) {
	donePath := test_env.SetPathWithCurrent(testdataBinPath)
	defer donePath()

	doneEnv := test_env.SetEnv("FAIL_WITH", "expect failure")
	defer doneEnv()

	Ffmpeg = newFfmpegCli()

	if got, want := Ffmpeg.err, ErrNoDependency; got != want {
		t.Errorf("Ffmpeg.err = %v, want: %v", got, want)
	}

	if Ffmpeg.IsInstalled() {
		t.Error("Ffmpeg should not be installed, but is found:", Ffmpeg)
	}

	if got, want := Ffmpeg.EncodeMp4("input", "output"), ErrNoDependency; !errors.Is(got, want) {
		t.Errorf("Ffmpge.EncodeMp4() = %v, want: %v", got, want)
	}

	if got, want := Ffmpeg.EncodeVideoThumbnail("input", "output", nil), ErrNoDependency; !errors.Is(got, want) {
		t.Errorf("Ffmpge.EncodeMp4() = %v, want: %v", got, want)
	}
}

func TestFfmpegIgnore(t *testing.T) {
	donePath := test_env.SetPathWithCurrent(testdataBinPath)
	defer donePath()

	doneEnv := test_env.SetEnv("PHOTOVIEW_DISABLE_VIDEO_ENCODING", "true")
	defer doneEnv()

	Ffmpeg = newFfmpegCli()

	if got, want := Ffmpeg.err, ErrDisabledFunction; got != want {
		t.Errorf("Ffmpeg.err = %v, want: %v", got, want)
	}

	if Ffmpeg.IsInstalled() {
		t.Error("Ffmpeg should be ignored (as it is disabled), but is initialized:", Ffmpeg)
	}

	if got, want := Ffmpeg.EncodeMp4("input", "output"), ErrDisabledFunction; !errors.Is(got, want) {
		t.Errorf("Ffmpge.EncodeMp4() = %v, want: %v", got, want)
	}

	if got, want := Ffmpeg.EncodeVideoThumbnail("input", "output", nil), ErrDisabledFunction; !errors.Is(got, want) {
		t.Errorf("Ffmpge.EncodeMp4() = %v, want: %v", got, want)
	}
}

func TestFfmpeg(t *testing.T) {
	done := test_env.SetPathWithCurrent(testdataBinPath)
	defer done()

	Ffmpeg = newFfmpegCli()

	if !Ffmpeg.IsInstalled() {
		t.Fatal("Ffmpeg should be installed")
	}

	t.Run("EncodeMp4Failed", func(t *testing.T) {
		doneEnv := test_env.SetEnv("FAIL_WITH", "expect failure")
		defer doneEnv()

		err := Ffmpeg.EncodeMp4("input", "output")
		if err == nil {
			t.Fatalf("Ffmpeg.EncodeMp4(...) = nil, should be an error.")
		}
		if got, want := err.Error(), `^encoding video with ".*/test_data/bin/ffmpeg" \[-i input -vcodec h264 .* output\] error: .*$`; !regexp.MustCompile(want).MatchString(got) {
			t.Errorf("Ffmpeg.EncodeMp4(...) = %q, should be as reg pattern %q", got, want)
		}
	})

	t.Run("EncodeMp4Succeeded", func(t *testing.T) {
		err := Ffmpeg.EncodeMp4("input", "output")
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
		doneEnv := test_env.SetEnv("FAIL_WITH", "expect failure")
		defer doneEnv()

		err := Ffmpeg.EncodeVideoThumbnail("input", "output", probeData)
		if err == nil {
			t.Fatalf("Ffmpeg.EncodeVideoThumbnail(...) = nil, should be an error.")
		}
		if got, want := err.Error(), `^encoding video thumbnail with ".*/test_data/bin/ffmpeg" \[-ss 2 -i input .* output\] error: .*$`; !regexp.MustCompile(want).MatchString(got) {
			t.Errorf("Ffmpeg.EncodeVideoThumbnail(...) = %q, should be as reg pattern %q", got, want)
		}
	})

	t.Run("EncodeVideoThumbnailSucceeded", func(t *testing.T) {
		err := Ffmpeg.EncodeVideoThumbnail("input", "output", probeData)
		if err != nil {
			t.Fatalf("Ffmpeg.EncodeVideoThumbnail(...) = %v, should be nil.", err)
		}
	})
}

func TestFfmpegWithHWAcc(t *testing.T) {
	doneCodec := test_env.SetEnv(utils.EnvVideoHardwareAcceleration.GetName(), "qsv")
	defer doneCodec()

	donePath := test_env.SetPathWithCurrent(testdataBinPath)
	defer donePath()

	Ffmpeg = newFfmpegCli()

	doneEnv := test_env.SetEnv("FAIL_WITH", "expect failure")
	defer doneEnv()

	err := Ffmpeg.EncodeMp4("input", "output")
	if err == nil {
		t.Fatalf("Ffmpeg.EncodeMp4(...) = nil, should be an error.")
	}
	if got, want := err.Error(), `^encoding video with ".*/test_data/bin/ffmpeg" \[-i input -vcodec h264_qsv .* output\] error: .*$`; !regexp.MustCompile(want).MatchString(got) {
		t.Errorf("Ffmpeg.EncodeMp4(...) = %q, should be as reg pattern %q", got, want)
	}
}

func TestFfmpegWithCustomCOdec(t *testing.T) {
	doneCodec := test_env.SetEnv(utils.EnvVideoHardwareAcceleration.GetName(), "_custom")
	defer doneCodec()

	donePath := test_env.SetPathWithCurrent(testdataBinPath)
	defer donePath()

	Ffmpeg = newFfmpegCli()

	doneEnv := test_env.SetEnv("FAIL_WITH", "expect failure")
	defer doneEnv()

	err := Ffmpeg.EncodeMp4("input", "output")
	if err == nil {
		t.Fatalf("Ffmpeg.EncodeMp4(...) = nil, should be an error.")
	}
	if got, want := err.Error(), `^encoding video with ".*/test_data/bin/ffmpeg" \[-i input -vcodec custom .* output\] error: .*$`; !regexp.MustCompile(want).MatchString(got) {
		t.Errorf("Ffmpeg.EncodeMp4(...) = %q, should be as reg pattern %q", got, want)
	}
}
