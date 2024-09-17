package executable_worker

func InitializeExecutableWorkers() {
	Magick = newMagickCli()
	Ffmpeg = newFfmpegCli()
}

var Magick *MagickCli = nil
var Ffmpeg *FfmpegCli = nil

type ExecutableWorker interface {
	Path() string
}
