package executable_worker

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"gopkg.in/vansante/go-ffprobe.v2"
)

func InitializeExecutableWorkers() {
	Magick = newMagickCli()
	Ffmpeg = newFfmpegCli()

	if err := SetFfprobePath(); err != nil {
		log.Println("ffprobe init fail:", err)
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

	log.Println("Found ffprobe:", path, "version:", strings.Split(string(version), "\n")[0])
	ffprobe.SetFFProbeBinPath(path)

	return nil
}
