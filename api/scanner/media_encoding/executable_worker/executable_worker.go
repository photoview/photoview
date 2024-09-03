package executable_worker

func InitializeExecutableWorkers() {
	MagickCli = newMagickWorker()
	FfmpegCli = newFfmpegWorker()
}

var MagickCli *MagickWorker = nil
var FfmpegCli *FfmpegWorker = nil

type ExecutableWorker interface {
	Path() string
}
