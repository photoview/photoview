package executable_worker

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/photoview/photoview/api/log"
	"gopkg.in/vansante/go-ffprobe.v2"
)

var ErrNoDependency = errors.New("dependency not found")
var ErrDisabledFunction = errors.New("function disabled")

// Initialize Initializes all workers. It returns a function to terminate workers, which should be called before the program closing.
func Initialize() func() {
	Magick = newMagickWand()
	Ffmpeg = newFfmpegCli()

	if err := SetFfprobePath(); err != nil {
		log.Error(nil, "Init ffprobe fail.", "error", err)
	}

	return func() {
		Magick.Terminate()
		Magick = nil
	}
}

var Magick *MagickWand = nil
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

	log.Info(nil, "Found ffprobe", "path", path, "version", strings.Split(string(version), "\n")[0])
	ffprobe.SetFFProbeBinPath(path)

	return nil
}
