package executable_worker

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/photoview/photoview/api/utils"
	"github.com/pkg/errors"
	"gopkg.in/vansante/go-ffprobe.v2"
)

func InitializeExecutableWorkers() {
	DarktableCli = newDarktableWorker()
	FfmpegCli = newFfmpegWorker()
}

var DarktableCli *DarktableWorker = nil
var FfmpegCli *FfmpegWorker = nil

type ExecutableWorker interface {
	Path() string
}

type DarktableWorker struct {
	path string
}

type FfmpegWorker struct {
	path string
}

func newDarktableWorker() *DarktableWorker {
	if utils.EnvDisableRawProcessing.GetBool() {
		log.Printf("Executable worker disabled (%s=1): darktable\n", utils.EnvDisableRawProcessing.GetName())
		return nil
	}

	path, err := exec.LookPath("darktable-cli")
	if err != nil {
		log.Println("Executable worker not found: darktable")
	} else {
		version, err := exec.Command(path, "--version").Output()
		if err != nil {
			log.Printf("Error getting version of darktable: %s\n", err)
			return nil
		}

		log.Printf("Found executable worker: darktable (%s)\n", strings.Split(string(version), "\n")[0])

		return &DarktableWorker{
			path: path,
		}
	}

	return nil
}

func newFfmpegWorker() *FfmpegWorker {
	if utils.EnvDisableVideoEncoding.GetBool() {
		log.Printf("Executable worker disabled (%s=1): ffmpeg\n", utils.EnvDisableVideoEncoding.GetName())
		return nil
	}

	path, err := exec.LookPath("ffmpeg")
	if err != nil {
		log.Println("Executable worker not found: ffmpeg")
	} else {
		version, err := exec.Command(path, "-version").Output()
		if err != nil {
			log.Printf("Error getting version of ffmpeg: %s\n", err)
			return nil
		}

		log.Printf("Found executable worker: ffmpeg (%s)\n", strings.Split(string(version), "\n")[0])

		return &FfmpegWorker{
			path: path,
		}
	}

	return nil
}

func (worker *DarktableWorker) IsInstalled() bool {
	return worker != nil
}

func (worker *FfmpegWorker) IsInstalled() bool {
	return worker != nil
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

func (worker *FfmpegWorker) EncodeVideoThumbnail(inputPath string, outputPath string, probeData *ffprobe.ProbeData) error {

	thumbnailOffsetSeconds := fmt.Sprintf("%d", int(probeData.Format.DurationSeconds*0.25))

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
