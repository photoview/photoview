package executable_worker

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/photoview/photoview/api/utils"
)

type MagickCli struct {
	path string
}

func newMagickCli() *MagickCli {
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

	return &MagickCli{
		path: path,
	}
}

func (cli *MagickCli) IsInstalled() bool {
	return cli != nil
}

func (cli *MagickCli) EncodeJpeg(inputPath string, outputPath string, jpegQuality int) error {
	args := []string{
		"convert",
		inputPath,
		"-quality", fmt.Sprintf("%d", jpegQuality),
		outputPath,
	}

	cmd := exec.Command(cli.path, args...)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("encoding image with \"%s %v\" error: %w", cli.path, args, err)
	}

	return nil
}
