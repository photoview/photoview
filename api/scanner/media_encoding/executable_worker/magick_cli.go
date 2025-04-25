package executable_worker

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/kkovaletp/photoview/api/log"
	"github.com/kkovaletp/photoview/api/utils"
)

type MagickCli struct {
	path string
	err  error
}

func newMagickCli() *MagickCli {
	if utils.EnvDisableRawProcessing.GetBool() {
		log.Warn("Executable magick worker disabled", utils.EnvDisableRawProcessing.GetName(), utils.EnvDisableRawProcessing.GetValue())
		return &MagickCli{
			err: ErrDisabledFunction,
		}
	}

	path, err := exec.LookPath("magick")
	if err != nil {
		log.Error("Executable magick worker not found")
		return &MagickCli{
			err: ErrNoDependency,
		}
	}

	version, err := exec.Command(path, "-version").Output()
	if err != nil {
		log.Error("Executable magick worker get version error", "error", err)
		return &MagickCli{
			err: ErrNoDependency,
		}
	}

	log.Info("Found magick executable worker", "version", strings.Split(string(version), "\n")[0])

	return &MagickCli{
		path: path,
	}
}

func (cli *MagickCli) IsInstalled() bool {
	return cli.err == nil
}

func (cli *MagickCli) EncodeJpeg(inputPath string, outputPath string, jpegQuality int) error {
	if cli.err != nil {
		return fmt.Errorf("encoding jpeg %q error: magick: %w", inputPath, cli.err)
	}

	args := []string{
		inputPath,
		"-auto-orient",
		"-quality", fmt.Sprintf("%d", jpegQuality),
		outputPath,
	}

	cmd := exec.Command(cli.path, args...)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("encoding image with \"%s %v\" error: %w", cli.path, args, err)
	}

	return nil
}
