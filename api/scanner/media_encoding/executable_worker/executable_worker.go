package executable_worker

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/photoview/photoview/api/utils"
	"github.com/pkg/errors"
	"gopkg.in/vansante/go-ffprobe.v2"
)

func InitializeExecutableWorkers() {
	MagickCli = newMagickWorker()
	FfmpegCli = newFfmpegWorker()
}

var MagickCli *MagickWorker = nil
var FfmpegCli *FfmpegWorker = nil

type ExecutableWorker interface {
	Path() string
}

type MagickWorker struct {
	path string
}

type FfmpegWorker struct {
	path    string
	encoder string
}

func newMagickWorker() *MagickWorker {
	if utils.EnvDisableRawProcessing.GetBool() {
		log.Printf("Executable worker disabled (%s=1): ImageMagick\n", utils.EnvDisableRawProcessing.GetName())
		return nil
	}

	path, err := exec.LookPath("convert")
	if err != nil {
		log.Println("Executable worker not found: ImageMagick convert")
	} else {
		version, err := exec.Command(path, "--version").Output()
		if err != nil {
			log.Printf("Error getting version of ImageMagick convert: %s\n", err)
			return nil
		}

		log.Printf("Found executable worker: ImageMagick convert (%s)\n", strings.Split(string(version), "\n")[0])

		return &MagickWorker{
			path: path,
		}
}

func newFfmpegWorker() *FfmpegWorker {
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
		log.Printf("Executable worker disabled (%s=1): ffmpeg\n", utils.EnvDisableVideoEncoding.GetName())
		return nil
	}

	path, err := exec.LookPath("ffmpeg")
	if err != nil {
		log.Println("Executable worker not found: ffmpeg")
		return nil
	}

	encoder := "h264"
	if e := utils.EnvVideoEncoder.GetValue(); e != "" {
		encoder = e
	}

	version, err := exec.Command(path, "-version").Output()
	if err != nil {
		log.Printf("Error getting version of ffmpeg: %s\n", err)
		return nil
	}

	log.Printf("Found executable worker: ffmpeg (%s), with encoder: %q\n", strings.Split(string(version), "\n")[0], encoder)

	return &FfmpegWorker{
		path:    path,
		encoder: encoder,
	}
}

func (worker *MagickWorker) IsInstalled() bool {
	return worker != nil
}

func (worker *FfmpegWorker) IsInstalled() bool {
	return worker != nil
}

func (worker *MagickWorker) EncodeJpeg(inputPath string, outputPath string, jpegQuality int) error {
	args := []string{
		inputPath,
		"-quality", fmt.Sprintf("%d", jpegQuality),
		outputPath,
	}

	cmd := exec.Command(worker.path, args...)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("encoding image with \"%s %v\" error: %w", worker.path, args, err)
	}

	return nil
}

func (worker *FfmpegWorker) EncodeMp4(inputPath string, outputPath string) error {
	args := []string{
		"-i",
		inputPath,
		"-vcodec", worker.encoder,
		"-acodec", "aac",
		"-vf", "scale='min(1080,iw)':'min(1080,ih)':force_original_aspect_ratio=decrease:force_divisible_by=2",
		"-movflags", "+faststart+use_metadata_tags",
		outputPath,
	}

	cmd := exec.Command(worker.path, args...)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("encoding video with ffmpeg %q encoder %q error: %w", worker.path, worker.encoder, err)
	}

	return nil
}

func (worker *FfmpegWorker) EncodeVideoThumbnail(inputPath string, outputPath string, probeData *ffprobe.ProbeData) error {

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
		return errors.Wrapf(err, "encoding video using: %s", worker.path)
	}

	return nil
}
