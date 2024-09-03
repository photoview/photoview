package executable_worker

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/photoview/photoview/api/utils"
)

type MagickWorker struct {
	path string
}

func newMagickWorker() *MagickWorker {
	if utils.EnvDisableRawProcessing.GetBool() {
		log.Printf("Executable worker disabled (%s=1): ImageMagick\n", utils.EnvDisableRawProcessing.GetName())
		return nil
	}

	path, err := exec.LookPath("magick")
	if err != nil {
		log.Println("Executable worker not found: magick")
		return nil
	}

	version, err := exec.Command(path, "-version").Output()
	if err != nil {
		log.Printf("Error getting version of magick: %s\n", err)
		return nil
	}

	log.Printf("Found executable worker: magick (%s)\n", strings.Split(string(version), "\n")[0])

	return &MagickWorker{
		path: path,
	}
}

func (worker *MagickWorker) IsInstalled() bool {
	return worker != nil
}

func (worker *MagickWorker) EncodeJpeg(inputPath string, outputPath string, jpegQuality int) error {
	args := []string{
		"convert",
		inputPath,
		"-quality", fmt.Sprintf("%d", jpegQuality),
		outputPath,
	}

	cmd := exec.Command(worker.path, args...)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("encoding image with %q %v error: %w", worker.path, args, err)
	}

	return nil
}
