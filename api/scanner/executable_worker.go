package scanner

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

var DarktableCli = newDarktableWorker()
var FfmpegCli = newFfmpegWorker()

type ExecutableWorker interface {
	Path() string
}

type DarktableWorker struct {
	path string
}

type FfmpegWorker struct {
	path string
}

func newDarktableWorker() DarktableWorker {
	path, err := exec.LookPath("darktable-cli")
	if err != nil {
		log.Println("Executable worker not found: darktable")
	} else {
		log.Println("Found executable worker: darktable")
	}

	return DarktableWorker{
		path: path,
	}
}

func newFfmpegWorker() FfmpegWorker {
	path, err := exec.LookPath("ffmpeg")
	if err != nil {
		log.Println("Executable worker not found: ffmpeg")
	} else {
		log.Println("Found executable worker: ffmpeg")
	}

	return FfmpegWorker{
		path: path,
	}
}

func (worker *DarktableWorker) IsInstalled() bool {
	return worker.path != ""
}

func (worker *FfmpegWorker) IsInstalled() bool {
	return worker.path != ""
}

func (worker *DarktableWorker) EncodeJpeg(inputPath string, outputPath string, jpegQuality int) error {
	tmpDir, err := ioutil.TempDir("/tmp", "photoview-darktable")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	args := []string{
		inputPath,
		outputPath,
		"--core",
		"--conf",
		fmt.Sprintf("plugins/imageio/format/jpeg/quality=%d", jpegQuality),
		"--configdir",
		tmpDir,
	}

	cmd := exec.Command(worker.path, args...)

	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "encoding image using: %s %v", worker.path, args)
	}

	return nil
}

func (worker *FfmpegWorker) EncodeMp4(inputPath string, outputPath string) error {
	args := []string{
		"-i",
		inputPath,
		"-vcodec", "h264",
		"-acodec", "aac",
		"-vf", "scale='min(1080,iw)':'min(1080,ih)':force_original_aspect_ratio=decrease",
		outputPath,
	}

	cmd := exec.Command(worker.path, args...)

	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "encoding video using: %s", worker.path)
	}

	return nil
}

func (worker *FfmpegWorker) EncodeVideoThumbnail(inputPath string, outputPath string, mediaData *EncodeMediaData) error {

	metadata, err := mediaData.VideoMetadata()
	if err != nil {
		return errors.Wrapf(err, "get metadata to encode video thumbnail (%s)", inputPath)
	}

	thumbnailOffsetSeconds := fmt.Sprintf("%d", int(metadata.Format.DurationSeconds*0.25))

	args := []string{
		"-i",
		inputPath,
		"-vframes", "1", // output one frame
		"-an", // disable audio
		"-vf", "scale='min(1024,iw)':'min(1024,ih)':force_original_aspect_ratio=decrease",
		"-ss", thumbnailOffsetSeconds, // grab frame at time offset
		outputPath,
	}

	cmd := exec.Command(worker.path, args...)

	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "encoding video using: %s", worker.path)
	}

	return nil
}
