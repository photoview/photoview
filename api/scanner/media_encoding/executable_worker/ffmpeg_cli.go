package executable_worker

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/photoview/photoview/api/log"
	"github.com/photoview/photoview/api/utils"
	"gopkg.in/vansante/go-ffprobe.v2"
)

const defaultCodec = "h264"

var hwAccToCodec = map[string]string{
	"qsv":   defaultCodec + "_qsv",
	"vaapi": defaultCodec + "_vaapi",
	"nvenc": defaultCodec + "_nvenc",
}

type FfmpegCli struct {
	path       string
	videoCodec string
}

func newFfmpegCli() *FfmpegCli {
	if utils.EnvDisableVideoEncoding.GetBool() {
		log.Warn("Executable ffmpeg worker disabled", utils.EnvDisableVideoEncoding.GetName(), utils.EnvDisableVideoEncoding.GetValue())
		return nil
	}

	path, err := exec.LookPath("ffmpeg")
	if err != nil {
		log.Error("Executable ffmpeg worker not found")
		return nil
	}

	version, err := exec.Command(path, "-version").Output()
	if err != nil {
		log.Error("Executable ffmpeg worker getting version error", "error", err)
		return nil
	}

	hwAcc := utils.EnvVideoHardwareAcceleration.GetValue()
	codec, ok := hwAccToCodec[hwAcc]
	if !ok {
		if strings.HasPrefix(hwAcc, "_") {
			// A secret way to set the codec directly.
			codec = hwAcc[1:]
		} else {
			codec = defaultCodec
		}
	}

	log.Info("Found executable worker: ffmpeg", "version", strings.Split(string(version), "\n")[0], "codec", codec)

	return &FfmpegCli{
		path:       path,
		videoCodec: codec,
	}
}

func (cli *FfmpegCli) IsInstalled() bool {
	return cli != nil
}

func (cli *FfmpegCli) EncodeMp4(inputPath string, outputPath string) error {
	if cli == nil {
		err := ErrNoDependency
		if utils.EnvDisableVideoEncoding.GetBool() {
			err = ErrDisabledFunction
		}
		return fmt.Errorf("encoding video %q error: ffmpeg: %w", inputPath, err)
	}

	args := []string{
		"-i",
		inputPath,
		"-vcodec", cli.videoCodec,
		"-acodec", "aac",
		"-vf", "scale='min(1080,iw)':'min(1080,ih)':force_original_aspect_ratio=decrease:force_divisible_by=2",
		"-movflags", "+faststart+use_metadata_tags",
		outputPath,
	}

	cmd := exec.Command(cli.path, args...)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("encoding video with %q %v error: %w", cli.path, args, err)
	}

	return nil
}

func (cli *FfmpegCli) EncodeVideoThumbnail(inputPath string, outputPath string, probeData *ffprobe.ProbeData) error {
	if cli == nil {
		err := ErrNoDependency
		if utils.EnvDisableVideoEncoding.GetBool() {
			err = ErrDisabledFunction
		}
		return fmt.Errorf("encoding video thumbnail %q error: ffmpeg: %w", inputPath, err)
	}

	thumbnailOffsetSeconds := fmt.Sprintf("%.f", probeData.Format.DurationSeconds*0.25)

	args := []string{
		"-ss", thumbnailOffsetSeconds, // grab frame at time offset
		"-i",
		inputPath,
		"-vframes", "1", // output one frame
		"-an", // disable audio
		"-vf", "scale='min(1024,iw)':'min(1024,ih)':force_original_aspect_ratio=decrease:force_divisible_by=2",
		outputPath,
	}

	cmd := exec.Command(cli.path, args...)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("encoding video thumbnail with %q %v error: %w", cli.path, args, err)
	}

	return nil
}
