package executable_worker

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/photoview/photoview/api/utils"
	"gopkg.in/vansante/go-ffprobe.v2"
)

type FfmpegCli struct {
	path       string
	videoCodec string
}

func newFfmpegCli() *FfmpegCli {
	if utils.EnvDisableVideoEncoding.GetBool() {
		log.Printf("Executable worker disabled (%s=%q): ffmpeg\n", utils.EnvDisableVideoEncoding.GetName(), utils.EnvDisableVideoEncoding.GetValue())
		return nil
	}

	path, err := exec.LookPath("ffmpeg")
	if err != nil {
		log.Println("Executable worker not found: ffmpeg")
		return nil
	}

	version, err := exec.Command(path, "-version").Output()
	if err != nil {
		log.Printf("Error getting version of ffmpeg: %s\n", err)
		return nil
	}

	codec := utils.EnvVideoEncoder.GetValue()
	fmt.Println("codec:", codec)
	if codec == "" {
		codec = "h264"
	}

	log.Printf("Found executable worker: ffmpeg (%s) with codec %q\n", strings.Split(string(version), "\n")[0], codec)

	return &FfmpegCli{
		path:       path,
		videoCodec: codec,
	}
}

func (worker *FfmpegCli) IsInstalled() bool {
	return worker != nil
}

func (worker *FfmpegCli) EncodeMp4(inputPath string, outputPath string) error {
	args := []string{
		"-i",
		inputPath,
		"-vcodec", worker.videoCodec,
		"-acodec", "aac",
		"-vf", "scale='min(1080,iw)':'min(1080,ih)':force_original_aspect_ratio=decrease:force_divisible_by=2",
		"-movflags", "+faststart+use_metadata_tags",
		outputPath,
	}

	cmd := exec.Command(worker.path, args...)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("encoding video with %q %v error: %w", worker.path, args, err)
	}

	return nil
}

func (worker *FfmpegCli) EncodeVideoThumbnail(inputPath string, outputPath string, probeData *ffprobe.ProbeData) error {

	thumbnailOffsetSeconds := fmt.Sprintf("%.f", probeData.Format.DurationSeconds*0.25)

	args := []string{
		"-ss", thumbnailOffsetSeconds, // grab frame at time offset
		"-i",
		inputPath,
		"-vframes", "1", // output one frame
		"-an", // disable audio
		"-vf", "scale='min(1024,iw)':'min(1024,ih)':force_original_aspect_ratio=decrease:force_divisible_by=2",
		outputPath,
	}

	cmd := exec.Command(worker.path, args...)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("encoding video thumbnail with %q %v error: %w", worker.path, args, err)
	}

	return nil
}
