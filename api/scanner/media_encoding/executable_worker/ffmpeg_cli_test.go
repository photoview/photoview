package executable_worker_test

import (
	"regexp"
	"testing"

	"github.com/photoview/photoview/api/scanner/media_encoding/executable_worker"
	"gopkg.in/vansante/go-ffprobe.v2"
)

func TestFfmpegWorkerNotExist(t *testing.T) {
	done := setPathWithCurrent()
	defer done()

	executable_worker.InitializeExecutableWorkers()

	if executable_worker.Ffmpeg.IsInstalled() {
		t.Error("Ffmpeg should not be installed, but is found:", executable_worker.Ffmpeg)
	}
}

func TestFfmpegWorkerIgnore(t *testing.T) {
	donePath := setPathWithCurrent("./testdata/bin")
	defer donePath()

	doneEnv := setEnv("PHOTOVIEW_DISABLE_VIDEO_ENCODING", "true")
	defer doneEnv()

	executable_worker.InitializeExecutableWorkers()

	if executable_worker.Ffmpeg.IsInstalled() {
		t.Error("Ffmpeg should not be installed, but is found:", executable_worker.Ffmpeg)
	}
}

func TestFfmpegWorker(t *testing.T) {
	done := setPathWithCurrent("./testdata/bin")
	defer done()

	executable_worker.InitializeExecutableWorkers()

	if !executable_worker.Ffmpeg.IsInstalled() {
		t.Error("Ffmpeg should be installed")
	}

	t.Run("EncodeMp4Failed", func(t *testing.T) {
		doneEnv := setEnv("FAIL_WITH", "expect failure")
		defer doneEnv()

		err := executable_worker.Ffmpeg.EncodeMp4("input_fail", "output")
		if err == nil {
			t.Fatalf("Ffmpeg.EncodeMp4(...) = nil, should be an error.")
		}
		if got, want := err.Error(), `^encoding video with ".*/testdata/bin/ffmpeg" .* error: .*$`; !regexp.MustCompile(want).MatchString(got) {
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
		doneEnv := setEnv("FAIL_WITH", "expect failure")
		defer doneEnv()

		err := executable_worker.Ffmpeg.EncodeVideoThumbnail("input_fail", "output", probeData)
		if err == nil {
			t.Fatalf("Ffmpeg.EncodeVideoThumbnail(...) = nil, should be an error.")
		}
		if got, want := err.Error(), `^encoding video thumbnail with ".*/testdata/bin/ffmpeg" .* error: .*$`; !regexp.MustCompile(want).MatchString(got) {
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
