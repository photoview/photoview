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
		if got, want := err.Error(), `^encoding video with ".*/test_data/mock_bin/ffmpeg" \[-i pipe:0 -vcodec h264 .* pipe:1\] error: .*$`; !regexp.MustCompile(want).MatchString(got) {
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
		if got, want := err.Error(), `^encoding video thumbnail with ".*/test_data/mock_bin/ffmpeg" \[-i pipe:0 -ss 2 .* pipe:1\] error: .*$`; !regexp.MustCompile(want).MatchString(got) {
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
	if got, want := err.Error(), `^encoding video with ".*/test_data/mock_bin/ffmpeg" \[-i pipe:0 -vcodec h264_qsv .* pipe:1\] error: .*$`; !regexp.MustCompile(want).MatchString(got) {
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
	if got, want := err.Error(), `^encoding video with ".*/test_data/mock_bin/ffmpeg" \[-i pipe:0 -vcodec custom .* pipe:1\] error: .*$`; !regexp.MustCompile(want).MatchString(got) {
		t.Errorf("Ffmpeg.EncodeMp4(...) = %q, should be as reg pattern %q", got, want)
	}
}

func TestFfmpegWithMemMapFs(t *testing.T) {
	SetPathWithCurrent(t, testdataBinPath)

	Ffmpeg = newFfmpegCli()

	if !Ffmpeg.IsInstalled() {
		t.Fatal("Ffmpeg should be installed")
	}

	t.Run("EncodeMp4WithMemMapFs", func(t *testing.T) {
		// Create an in-memory filesystem
		memFs := afero.NewMemMapFs()

		// Read the actual sample video file from disk
		sampleVideoPath := "test_data/sample_video.avi"
		osFs := afero.NewOsFs()
		videoData, err := afero.ReadFile(osFs, sampleVideoPath)
		if err != nil {
			t.Fatalf("Failed to read sample video file: %v", err)
		}

		// Copy the sample video to MemMapFs
		inputPath := "/test/input.avi"
		if err := memFs.MkdirAll("/test", 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := afero.WriteFile(memFs, inputPath, videoData, 0644); err != nil {
			t.Fatalf("Failed to write input file to MemMapFs: %v", err)
		}

		outputPath := "/test/output.mp4"

		// Run the encoding
		err = Ffmpeg.EncodeMp4(memFs, inputPath, outputPath)
		if err != nil {
			t.Fatalf("Ffmpeg.EncodeMp4(...) with MemMapFs = %v, should be nil.", err)
		}

		// Verify output file was created
		exists, err := afero.Exists(memFs, outputPath)
		if err != nil {
			t.Fatalf("Failed to check if output exists: %v", err)
		}
		if !exists {
			t.Errorf("Output file %q should exist in MemMapFs", outputPath)
		}

		// Verify output file has content
		outputData, err := afero.ReadFile(memFs, outputPath)
		if err != nil {
			t.Fatalf("Failed to read output file: %v", err)
		}
		if len(outputData) == 0 {
			t.Error("Output file should have content")
		}

		t.Logf("Encoded video: input size=%d bytes, output size=%d bytes", len(videoData), len(outputData))
	})

	t.Run("EncodeVideoThumbnailWithMemMapFs", func(t *testing.T) {
		// Create an in-memory filesystem
		memFs := afero.NewMemMapFs()

		// Read the actual sample video file from disk
		sampleVideoPath := "test_data/sample_video.avi"
		osFs := afero.NewOsFs()
		videoData, err := afero.ReadFile(osFs, sampleVideoPath)
		if err != nil {
			t.Fatalf("Failed to read sample video file: %v", err)
		}

		// Copy the sample video to MemMapFs
		inputPath := "/test/video.avi"
		if err := memFs.MkdirAll("/test", 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := afero.WriteFile(memFs, inputPath, videoData, 0644); err != nil {
			t.Fatalf("Failed to write input file to MemMapFs: %v", err)
		}

		outputPath := "/test/thumbnail.jpg"
		probeData := &ffprobe.ProbeData{
			Format: &ffprobe.Format{
				DurationSeconds: 30,
			},
		}

		// Run the thumbnail encoding
		err = Ffmpeg.EncodeVideoThumbnail(memFs, inputPath, outputPath, probeData)
		if err != nil {
			t.Fatalf("Ffmpeg.EncodeVideoThumbnail(...) with MemMapFs = %v, should be nil.", err)
		}

		// Verify output file was created
		exists, err := afero.Exists(memFs, outputPath)
		if err != nil {
			t.Fatalf("Failed to check if output exists: %v", err)
		}
		if !exists {
			t.Errorf("Output file %q should exist in MemMapFs", outputPath)
		}

		// Verify output file has content
		outputData, err := afero.ReadFile(memFs, outputPath)
		if err != nil {
			t.Fatalf("Failed to read output file: %v", err)
		}
		if len(outputData) == 0 {
			t.Error("Output file should have content")
		}

		t.Logf("Generated thumbnail: input size=%d bytes, output size=%d bytes", len(videoData), len(outputData))
	})

	t.Run("EncodeMp4WithMemMapFsInputNotFound", func(t *testing.T) {
		memFs := afero.NewMemMapFs()

		err := Ffmpeg.EncodeMp4(memFs, "/nonexistent/input.mp4", "/test/output.mp4")
		if err == nil {
			t.Fatal("Ffmpeg.EncodeMp4(...) should fail with non-existent input file")
		}
		if got, want := err.Error(), "opening input file"; !regexp.MustCompile(want).MatchString(got) {
			t.Errorf("Error should mention input file opening, got: %q", got)
		}
	})
}
