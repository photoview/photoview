package scanner

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/viktorstrate/photoview/api/graphql/models"
	"github.com/viktorstrate/photoview/api/utils"
	"gopkg.in/vansante/go-ffprobe.v2"
)

func processVideo(tx *sql.Tx, mediaData *EncodeMediaData, videoCachePath *string) (bool, error) {
	video := mediaData.media
	didProcess := false

	log.Printf("Processing video: %s", video.Path)

	mediaUrlFromDB, err := makePhotoURLChecker(tx, video.MediaID)
	if err != nil {
		return false, err
	}

	videoWebURL, err := mediaUrlFromDB(models.VideoWeb)
	if err != nil {
		return false, errors.Wrap(err, "error processing video web-format")
	}

	if videoWebURL == nil {
		web_video_name := fmt.Sprintf("web_video_%s_%s", path.Base(video.Path), utils.GenerateToken())
		web_video_name = strings.ReplaceAll(web_video_name, ".", "_")
		web_video_name = strings.ReplaceAll(web_video_name, " ", "_")
		web_video_name = web_video_name + ".mp4"

		webVideoPath := path.Join(*videoCachePath, web_video_name)

		err = FfmpegCli.EncodeMp4(video.Path, webVideoPath)
		if err != nil {
			return false, errors.Wrapf(err, "could not encode mp4 video (%s)", video.Path)
		}

		webMetadata, err := readVideoMetadata(webVideoPath)
		if err != nil {
			return false, errors.Wrapf(err, "failed to read metadata for encoded web-video (%s)", video.Title)
		}

		_, err = tx.Exec("INSERT INTO media_url (media_id, media_name, width, height, purpose, content_type) VALUES (?, ?, ?, ?, ?, ?)",
			video.MediaID, web_video_name, webMetadata.Width, webMetadata.Height, models.VideoWeb, "video/mp4")
		if err != nil {
			return false, errors.Wrapf(err, "failed to insert encoded web-video into database (%s)", video.Title)
		}
	}

	// TODO: Process video thumbnail

	return didProcess, nil
}

func (enc *EncodeMediaData) VideoMetadata() (*ffprobe.Stream, error) {

	if enc._videoMetadata != nil {
		return enc._videoMetadata, nil
	}

	metadata, err := readVideoMetadata(enc.media.Path)
	if err != nil {
		return nil, err
	}

	enc._videoMetadata = metadata
	return metadata, nil
}

func readVideoMetadata(videoPath string) (*ffprobe.Stream, error) {
	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()

	data, err := ffprobe.ProbeURL(ctx, videoPath)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read video metadata (%s)", path.Base(videoPath))
	}

	stream := data.FirstVideoStream()
	if stream == nil {
		return nil, errors.Wrapf(err, "could not get stream from file metadata (%s)", path.Base(videoPath))
	}

	return stream, nil
}
