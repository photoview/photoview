package scanner

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/photoview/photoview/api/graphql/models"
)

func ScanVideoMetadata(tx *sql.Tx, video *models.Media) error {

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

	result, err := tx.Exec("INSERT INTO video_metadata (width, height, duration, codec, framerate, bitrate, color_profile, audio) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", stream.Width, stream.Height, data.Format.DurationSeconds, stream.CodecLongName, framerate, stream.BitRate, stream.Profile, audioText)
	if err != nil {
		return errors.Wrapf(err, "failed to insert video metadata into database (%s)", video.Title)
	}

	metadata_id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	if _, err = tx.Exec("UPDATE media SET video_metadata_id = ? WHERE media_id = ?", metadata_id, video.MediaID); err != nil {
		return err
	}

	return nil
}
