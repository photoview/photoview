package scanner_tasks

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/scanner_task"
	"github.com/photoview/photoview/api/scanner/scanner_tasks/processing_tasks"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"gopkg.in/vansante/go-ffprobe.v2"
	"gorm.io/gorm"
)

type VideoMetadataTask struct {
	scanner_task.ScannerTaskBase
}

func (t VideoMetadataTask) AfterMediaFound(ctx scanner_task.TaskContext, media *models.Media, newMedia bool) error {
	db := ctx.GetDB()
	fs := ctx.GetFileFS()

	if !newMedia || media.Type != models.MediaTypeVideo {
		return nil
	}

	err := scanVideoMetadata(db, fs, media)
	if err != nil {
		log.Printf("WARN: ScanVideoMetadata for %s failed: %s\n", media.Title, err)
	}

	return nil
}

func scanVideoMetadata(tx *gorm.DB, fs afero.Fs, video *models.Media) error {
	localPath, err := video.GetLocalPath(fs)
	if err != nil {
		return errors.Wrapf(err, "could not get local path for video (%s)", video.Path)
	}

	data, err := processing_tasks.ReadVideoMetadata(*localPath)
	if err != nil {
		return errors.Wrapf(err, "scan video metadata failed (%s)", video.Title)
	}

	stream := data.FirstVideoStream()
	if stream == nil {
		return errors.New(fmt.Sprintf("could not get video stream from metadata (%s)", video.Path))
	}

	audio := data.FirstAudioStream()
	var audioText string
	if audio == nil {
		audioText = "No audio"
	} else {
		switch audio.Channels {
		case 0:
			audioText = "No audio"
		case 1:
			audioText = "Mono audio"
		case 2:
			audioText = "Stereo audio"
		default:
			audioText = fmt.Sprintf("Audio (%d channels)", audio.Channels)
		}
	}

	framerate := getFrameRate(stream)

	videoMetadata := models.VideoMetadata{
		Width:        stream.Width,
		Height:       stream.Height,
		Duration:     data.Format.DurationSeconds,
		Codec:        &stream.CodecLongName,
		Framerate:    framerate,
		Bitrate:      &stream.BitRate,
		ColorProfile: &stream.Profile,
		Audio:        &audioText,
	}

	video.VideoMetadata = &videoMetadata

	if err := tx.Save(video).Error; err != nil {
		return errors.Wrapf(err, "failed to add video metadata to database (%s)", video.Title)
	}

	return nil
}

func getFrameRate(stream *ffprobe.Stream) *float64 {
	var framerate *float64 = nil
	if stream.AvgFrameRate != "" {
		parts := strings.Split(stream.AvgFrameRate, "/")
		if len(parts) == 2 {
			if numerator, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
				if denominator, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
					result := float64(numerator) / float64(denominator)
					framerate = &result
				}
			}
		}
	}
	return framerate
}
