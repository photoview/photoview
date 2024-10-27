package executable_worker_test

import (
	"regexp"
	"testing"

	"github.com/photoview/photoview/api/scanner/media_encoding/executable_worker"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/photoview/photoview/api/utils"
	"gopkg.in/vansante/go-ffprobe.v2"
)

func TestFfmpegNotExist(t *testing.T) {
	done := test_utils.SetPathWithCurrent()
	defer done()

	executable_worker.InitializeExecutableWorkers()

	if executable_worker.Ffmpeg.IsInstalled() {
		t.Error("Ffmpeg should not be installed, but is found:", executable_worker.Ffmpeg)
	}
}

func TestFfmpegIgnore(t *testing.T) {
	donePath := test_utils.SetPathWithCurrent(testdataBinPath)
	defer donePath()

	doneEnv := test_utils.SetEnv("PHOTOVIEW_DISABLE_VIDEO_ENCODING", "true")
	defer doneEnv()

	executable_worker.InitializeExecutableWorkers()

	if executable_worker.Ffmpeg.IsInstalled() {
		t.Error("Ffmpeg should be ignored (as it is disabled), but is initialized:", executable_worker.Ffmpeg)
	}
}

func TestFfmpeg(t *testing.T) {
	done := test_utils.SetPathWithCurrent(testdataBinPath)
	defer done()

	executable_worker.InitializeExecutableWorkers()

	if !executable_worker.Ffmpeg.IsInstalled() {
		t.Fatal("Ffmpeg should be installed")
	}

	t.Run("EncodeMp4Failed", func(t *testing.T) {
		doneEnv := test_utils.SetEnv("FAIL_WITH", "expect failure")
		defer doneEnv()

		err := executable_worker.Ffmpeg.EncodeMp4("input", "output")
		if err == nil {
			t.Fatalf("Ffmpeg.EncodeMp4(...) = nil, should be an error.")
		}
		if got, want := err.Error(), `^encoding video with ".*/testdata/bin/ffmpeg" \[-i input -vcodec h264 .* output\] error: .*$`; !regexp.MustCompile(want).MatchString(got) {
			t.Errorf("Ffmpeg.EncodeMp4(...) = %q, should be as reg pattern %q", got, want)
		}
	})

	t.Run("EncodeMp4Succeeded", func(t *testing.T) {
		err := executable_worker.Ffmpeg.EncodeMp4("input", "output")
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
		doneEnv := test_utils.SetEnv("FAIL_WITH", "expect failure")
		defer doneEnv()

		err := executable_worker.Ffmpeg.EncodeVideoThumbnail("input", "output", probeData)
		if err == nil {
			t.Fatalf("Ffmpeg.EncodeVideoThumbnail(...) = nil, should be an error.")
		}
		if got, want := err.Error(), `^encoding video thumbnail with ".*/testdata/bin/ffmpeg" \[-ss 2 -i input .* output\] error: .*$`; !regexp.MustCompile(want).MatchString(got) {
			t.Errorf("Ffmpeg.EncodeVideoThumbnail(...) = %q, should be as reg pattern %q", got, want)
		}
	})

	t.Run("EncodeVideoThumbnailSucceeded", func(t *testing.T) {
		err := executable_worker.Ffmpeg.EncodeVideoThumbnail("input", "output", probeData)
		if err != nil {
			t.Fatalf("Ffmpeg.EncodeVideoThumbnail(...) = %v, should be nil.", err)
		}
	})
}

func TestFfmpegWithHWAcc(t *testing.T) {
	doneCodec := test_utils.SetEnv(utils.EnvVideoHardwareAcceleration.GetName(), "qsv")
	defer doneCodec()

	donePath := test_utils.SetPathWithCurrent(testdataBinPath)
	defer donePath()

	executable_worker.InitializeExecutableWorkers()

	doneEnv := test_utils.SetEnv("FAIL_WITH", "expect failure")
	defer doneEnv()

	err := executable_worker.Ffmpeg.EncodeMp4("input", "output")
	if err == nil {
		t.Fatalf("Ffmpeg.EncodeMp4(...) = nil, should be an error.")
	}
	if got, want := err.Error(), `^encoding video with ".*/testdata/bin/ffmpeg" \[-i input -vcodec h264_qsv .* output\] error: .*$`; !regexp.MustCompile(want).MatchString(got) {
		t.Errorf("Ffmpeg.EncodeMp4(...) = %q, should be as reg pattern %q", got, want)
	}
}

func TestFfmpegWithCustomCOdec(t *testing.T) {
	doneCodec := test_utils.SetEnv(utils.EnvVideoHardwareAcceleration.GetName(), "_custom")
	defer doneCodec()

	donePath := test_utils.SetPathWithCurrent(testdataBinPath)
	defer donePath()

	executable_worker.InitializeExecutableWorkers()

	doneEnv := test_utils.SetEnv("FAIL_WITH", "expect failure")
	defer doneEnv()

	err := executable_worker.Ffmpeg.EncodeMp4("input", "output")
	if err == nil {
		t.Fatalf("Ffmpeg.EncodeMp4(...) = nil, should be an error.")
	}
	if got, want := err.Error(), `^encoding video with ".*/testdata/bin/ffmpeg" \[-i input -vcodec custom .* output\] error: .*$`; !regexp.MustCompile(want).MatchString(got) {
		t.Errorf("Ffmpeg.EncodeMp4(...) = %q, should be as reg pattern %q", got, want)
	}
}
