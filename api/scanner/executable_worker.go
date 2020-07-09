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
		outputPath,
	}

	cmd := exec.Command(worker.path, args...)

	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "encoding video using: %s", worker.path)
	}

	return nil
}
