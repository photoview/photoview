package scanner

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/viktorstrate/photoview/api/graphql/models"
	"github.com/viktorstrate/photoview/api/utils"
	"gopkg.in/vansante/go-ffprobe.v2"
	"gorm.io/gorm"
)

func processVideo(tx *gorm.DB, mediaData *EncodeMediaData, videoCachePath *string) (bool, error) {
	video := mediaData.media
	didProcess := false

	log.Printf("Processing video: %s", video.Path)

	mediaUrlFromDB, err := makePhotoURLChecker(tx, video.ID)
	if err != nil {
		return false, err
	}

	videoWebURL, err := mediaUrlFromDB(models.VideoWeb)
	if err != nil {
		return false, errors.Wrap(err, "error processing video web-format")
	}

	videoThumbnailURL, err := mediaUrlFromDB(models.VideoThumbnail)
	if err != nil {
		return false, errors.Wrap(err, "error processing video thumbnail")
	}

	if videoWebURL == nil {
		didProcess = true

		web_video_name := fmt.Sprintf("web_video_%s_%s", path.Base(video.Path), utils.GenerateToken())
		web_video_name = strings.ReplaceAll(web_video_name, ".", "_")
		web_video_name = strings.ReplaceAll(web_video_name, " ", "_")
		web_video_name = web_video_name + ".mp4"

		webVideoPath := path.Join(*videoCachePath, web_video_name)

		err = FfmpegCli.EncodeMp4(video.Path, webVideoPath)
		if err != nil {
			return false, errors.Wrapf(err, "could not encode mp4 video (%s)", video.Path)
		}

		webMetadata, err := readVideoStreamMetadata(webVideoPath)
		if err != nil {
			return false, errors.Wrapf(err, "failed to read metadata for encoded web-video (%s)", video.Title)
		}

		fileStats, err := os.Stat(webVideoPath)
		if err != nil {
			return false, errors.Wrap(err, "reading file stats of web-optimized video")
		}

		_, err = tx.Exec("INSERT INTO media_url (media_id, media_name, width, height, purpose, content_type, file_size) VALUES (?, ?, ?, ?, ?, ?, ?)",
			video.MediaID, web_video_name, webMetadata.Width, webMetadata.Height, models.VideoWeb, "video/mp4", fileStats.Size())
		if err != nil {
			return false, errors.Wrapf(err, "failed to insert encoded web-video into database (%s)", video.Title)
		}
	}

	if videoThumbnailURL == nil {
		didProcess = true

		video_thumb_name := fmt.Sprintf("video_thumb_%s_%s", path.Base(video.Path), utils.GenerateToken())
		video_thumb_name = strings.ReplaceAll(video_thumb_name, ".", "_")
		video_thumb_name = strings.ReplaceAll(video_thumb_name, " ", "_")
		video_thumb_name = video_thumb_name + ".jpg"

		thumbImagePath := path.Join(*videoCachePath, video_thumb_name)

		err = FfmpegCli.EncodeVideoThumbnail(video.Path, thumbImagePath, mediaData)
		if err != nil {
			return false, errors.Wrapf(err, "failed to generate thumbnail for video (%s)", video.Title)
		}

		thumbDimensions, err := GetPhotoDimensions(thumbImagePath)
		if err != nil {
			return false, errors.Wrap(err, "get dimensions of video thumbnail image")
		}

		fileStats, err := os.Stat(thumbImagePath)
		if err != nil {
			return false, errors.Wrap(err, "reading file stats of video thumbnail")
		}

		_, err = tx.Exec("INSERT INTO media_url (media_id, media_name, width, height, purpose, content_type, file_size) VALUES (?, ?, ?, ?, ?, ?, ?)",
			video.MediaID, video_thumb_name, thumbDimensions.Width, thumbDimensions.Height, models.VideoThumbnail, "image/jpeg", fileStats.Size())
		if err != nil {
			return false, errors.Wrapf(err, "failed to insert video thumbnail image into database (%s)", video.Title)
		}
	}

	return didProcess, nil
}

func (enc *EncodeMediaData) VideoMetadata() (*ffprobe.ProbeData, error) {

	if enc._videoMetadata != nil {
		return enc._videoMetadata, nil
	}

	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()
	data, err := ffprobe.ProbeURL(ctx, enc.media.Path)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read video metadata (%s)", enc.media.Title)
	}

	enc._videoMetadata = data
	return enc._videoMetadata, nil
}

func readVideoMetadata(videoPath string) (*ffprobe.ProbeData, error) {
	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()

	data, err := ffprobe.ProbeURL(ctx, videoPath)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read video metadata (%s)", path.Base(videoPath))
	}

	return data, nil
}

func readVideoStreamMetadata(videoPath string) (*ffprobe.Stream, error) {
	data, err := readVideoMetadata(videoPath)
	if err != nil {
		return nil, errors.Wrap(err, "read video stream metadata")
	}

	stream := data.FirstVideoStream()
	if stream == nil {
		return nil, errors.Wrapf(err, "could not get stream from file metadata (%s)", path.Base(videoPath))
	}

	return stream, nil
}
