package scanner

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func ScanVideoMetadata(tx *gorm.DB, video *models.Media) error {

	data, err := readVideoMetadata(video.Path)
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
