package executable_worker

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/kkovaletp/photoview/api/log"
	"gopkg.in/vansante/go-ffprobe.v2"
)

var ErrNoDependency = errors.New("dependency not found")
var ErrDisabledFunction = errors.New("function disabled")

func init() {
	Magick = newMagickCli()
	Ffmpeg = newFfmpegCli()

	if err := SetFfprobePath(); err != nil {
		log.Error("Init ffprobe fail.", "error", err)
	}
}

var Magick *MagickCli = nil
var Ffmpeg *FfmpegCli = nil

type ExecutableWorker interface {
	Path() string
}

func SetFfprobePath() error {
	path, err := exec.LookPath("ffprobe")
	if err != nil {
		return fmt.Errorf("Executable ffprobe not found: %w", err)
	}

	version, err := exec.Command(path, "-version").Output()
	if err != nil {
		return fmt.Errorf("Executable ffprobe(%q) not executable: %w", path, err)
	}

	log.Info("Found ffprobe", "path", path, "version", strings.Split(string(version), "\n")[0])
	ffprobe.SetFFProbeBinPath(path)

	return nil
}
