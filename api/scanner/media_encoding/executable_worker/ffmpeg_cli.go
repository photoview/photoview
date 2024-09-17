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
	path string
}

func newFfmpegCli() *FfmpegCli {
	if path, err := exec.LookPath("ffprobe"); err == nil {
		if version, err := exec.Command(path, "-version").Output(); err == nil {
			log.Println("Found ffprobe:", path, "version:", strings.Split(string(version), "\n")[0])
			ffprobe.SetFFProbeBinPath(path)
		} else {
			log.Println("Executable ffprobe not executable:", path)
		}
	} else {
		log.Println("Executable ffprobe not found")
	}

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

	log.Printf("Found executable worker: ffmpeg (%s)\n", strings.Split(string(version), "\n")[0])

	return &FfmpegCli{
		path: path,
	}
}

func (worker *FfmpegCli) IsInstalled() bool {
	return worker != nil
}

func (worker *FfmpegCli) EncodeMp4(inputPath string, outputPath string) error {
	args := []string{
		"-i",
		inputPath,
		"-vcodec", "h264",
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

	thumbnailOffsetSeconds := fmt.Sprintf("%d", int(probeData.Format.DurationSeconds*0.25))

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
