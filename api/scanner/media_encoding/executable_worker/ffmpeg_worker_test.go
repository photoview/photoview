package executable_worker_test

import (
	"os"
	"regexp"
	"testing"

	"github.com/photoview/photoview/api/scanner/media_encoding/executable_worker"
	"gopkg.in/vansante/go-ffprobe.v2"
)

func TestFfmpegWorkerNotExist(t *testing.T) {
	done := setPathWithCurrent()
	defer done()

	executable_worker.InitializeExecutableWorkers()

	if executable_worker.FfmpegCli.IsInstalled() {
		t.Error("FfmpegCli should not be installed, but is found:", executable_worker.FfmpegCli)
	}
}

func TestFfmpegWorkerIgnore(t *testing.T) {
	done := setPathWithCurrent("./testdata/bin")
	defer done()

	envKey := "PHOTOVIEW_DISABLE_VIDEO_ENCODING"
	org := os.Getenv(envKey)
	os.Setenv(envKey, "true")
	defer os.Setenv(envKey, org)

	executable_worker.InitializeExecutableWorkers()

	if executable_worker.FfmpegCli.IsInstalled() {
		t.Error("FfmpegCli should not be installed, but is found:", executable_worker.FfmpegCli)
	}
}

func TestFfmpegWorker(t *testing.T) {
	done := setPathWithCurrent("./testdata/bin")
	defer done()

	executable_worker.InitializeExecutableWorkers()

	if !executable_worker.FfmpegCli.IsInstalled() {
		t.Error("FfmpegCli should be installed")
	}

	t.Run("EncodeMp4Failed", func(t *testing.T) {
		err := executable_worker.FfmpegCli.EncodeMp4("input_fail", "output")
		if err == nil {
			t.Fatalf("FfmpegCli.EncodeMp4(\"input_fail\", \"output\") = nil, should be an error.")
		}
		if got, want := err.Error(), `^encoding video with ".*/testdata/bin/ffmpeg" .* error: .*$`; !regexp.MustCompile(want).MatchString(got) {
			t.Errorf("FfmpegCli.EncodeMp4(\"input_fail\", \"output\") = %q, should be as reg pattern %q", got, want)
		}
	})

	t.Run("EncodeMp4Succeeded", func(t *testing.T) {
		err := executable_worker.FfmpegCli.EncodeMp4("input", "output")
		if err != nil {
			t.Fatalf("FfmpegCli.EncodeMp4(\"input\", \"output\") = %v, should be nil.", err)
		}
	})

	probeData := &ffprobe.ProbeData{
		Format: &ffprobe.Format{
			DurationSeconds: 10,
		},
	}
	t.Run("EncodeVideoThumbnailMp4Failed", func(t *testing.T) {
		err := executable_worker.FfmpegCli.EncodeVideoThumbnail("input_fail", "output", probeData)
		if err == nil {
			t.Fatalf("FfmpegCli.EncodeVideoThumbnail(\"input_fail\", \"output\") = nil, should be an error.")
		}
		if got, want := err.Error(), `^encoding video thumbnail with ".*/testdata/bin/ffmpeg" .* error: .*$`; !regexp.MustCompile(want).MatchString(got) {
			t.Errorf("FfmpegCli.EncodeVideoThumbnail(\"input_fail\", \"output\") = %q, should be as reg pattern %q", got, want)
		}
	})

	t.Run("EncodeVideoThumbnailSucceeded", func(t *testing.T) {
		err := executable_worker.FfmpegCli.EncodeVideoThumbnail("input", "output", probeData)
		if err != nil {
			t.Fatalf("FfmpegCli.EncodeVideoThumbnail(\"input\", \"output\") = %v, should be nil.", err)
		}
	})
}
