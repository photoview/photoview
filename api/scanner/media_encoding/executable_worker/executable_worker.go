package executable_worker

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"bytes"


	"github.com/photoview/photoview/api/utils"
	"github.com/pkg/errors"
	"gopkg.in/vansante/go-ffprobe.v2"
)

func InitializeExecutableWorkers() {
	CustomRawCli = newCustomRawWorker()
	FfmpegCli = newFfmpegWorker()
}

var CustomRawCli *CustomRawWorker = nil
var FfmpegCli *FfmpegWorker = nil

type ExecutableWorker interface {
	Path() string
}

type FfmpegWorker struct {
	path string
}

type CustomRawWorker struct {
	path string
	args string
}

type RawArgs struct {
    InputPath string
    OutputPath string
	Quality int
}


func newCustomRawWorker() *CustomRawWorker {
	if utils.EnvDisableRawProcessing.GetBool() {
		log.Printf("Executable worker disabled (%s=1) \n", utils.EnvDisableRawProcessing.GetName())
		return nil
	}

	var _check = utils.EnvRawProcessingCheck.GetValue()
	var _exec = utils.EnvRawProcessing.GetValue()
	var _args = utils.EnvRawProcessingArgs.GetValue()

	path, err := exec.LookPath(_exec)
	if err != nil {
		log.Println("Executable worker not found: %s",_exec)
	} else {
		version, err := exec.Command(path, _check).Output()
		if err != nil {
			log.Printf("Error getting version of %s: %s\n",_exec,_check, err)
			return nil
		}

		log.Printf("Found executable worker: %s (%s)\n", _exec, strings.Split(string(version), "\n")[0])

		return &CustomRawWorker{
			path: path,
			args: _args,
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

func (worker *CustomRawWorker) IsInstalled() bool {
	return worker != nil
}

func (worker *FfmpegWorker) IsInstalled() bool {
	return worker != nil
}

func (worker *CustomRawWorker) EncodeJpeg(inputPath string, outputPath string, jpegQuality int) error {
	tmpDir, err := ioutil.TempDir("/tmp", "photoview-convert")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	tmpl, err := template.New("RawArgs").Parse(worker.args)
	if err != nil { 
		log.Fatal(err)
    }
	_args := RawArgs{inputPath,outputPath,jpegQuality}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf,_args)
	if err != nil { 
        log.Fatal(err)
    }
	args := strings.Split(buf.String()," ")

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
		"-vf", "scale='min(1080,iw)':'min(1080,ih)':force_original_aspect_ratio=decrease:force_divisible_by=2",
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
		"-ss", thumbnailOffsetSeconds, // grab frame at time offset
		"-i",
		inputPath,
		"-vframes", "1", // output one frame
		"-an", // disable audio
		"-vf", "scale='min(1024,iw)':'min(1024,ih)':force_original_aspect_ratio=decrease:force_divisible_by=2",
		outputPath,
	}

	cmd := exec.Command(worker.path, args...)

	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "encoding video using: %s", worker.path)
	}

	return nil
}
