package scanner

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/image_helpers"
	"github.com/photoview/photoview/api/utils"
	"github.com/pkg/errors"
	"gopkg.in/vansante/go-ffprobe.v2"
	"gorm.io/gorm"
)

func processVideo(tx *gorm.DB, mediaData *EncodeMediaData, videoCachePath *string) (bool, error) {
	video := mediaData.media
	didProcess := false

	log.Printf("Processing video: %s", video.Path)

	mediaURLFromDB := makePhotoURLChecker(tx, video.ID)

	videoOriginalURL, err := mediaURLFromDB(models.MediaOriginal)
	if err != nil {
		return false, errors.Wrap(err, "error processing video original format")
	}

	videoWebURL, err := mediaURLFromDB(models.VideoWeb)
	if err != nil {
		return false, errors.Wrap(err, "error processing video web-format")
	}

	videoThumbnailURL, err := mediaURLFromDB(models.VideoThumbnail)
	if err != nil {
		return false, errors.Wrap(err, "error processing video thumbnail")
	}

	videoType, err := mediaData.ContentType()
	if err != nil {
		return false, errors.Wrap(err, "error getting video content type")
	}

	if videoOriginalURL == nil && videoType.isWebCompatible() {
		didProcess = true

		origVideoPath := video.Path
		videoMediaName := generateUniqueMediaName(video.Path)

		webMetadata, err := readVideoStreamMetadata(origVideoPath)
		if err != nil {
			return false, errors.Wrapf(err, "failed to read metadata for original video (%s)", video.Title)
		}

		fileStats, err := os.Stat(origVideoPath)
		if err != nil {
			return false, errors.Wrap(err, "reading file stats of original video")
		}

		mediaURL := models.MediaURL{
			MediaID:     video.ID,
			MediaName:   videoMediaName,
			Width:       webMetadata.Width,
			Height:      webMetadata.Height,
			Purpose:     models.MediaOriginal,
			ContentType: string(*videoType),
			FileSize:    fileStats.Size(),
		}

		if err := tx.Create(&mediaURL).Error; err != nil {
			return false, errors.Wrapf(err, "failed to insert original video into database (%s)", video.Title)
		}

	}

	if videoWebURL == nil && !videoType.isWebCompatible() {
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

		mediaURL := models.MediaURL{
			MediaID:     video.ID,
			MediaName:   web_video_name,
			Width:       webMetadata.Width,
			Height:      webMetadata.Height,
			Purpose:     models.VideoWeb,
			ContentType: "video/mp4",
			FileSize:    fileStats.Size(),
		}

		if err := tx.Create(&mediaURL).Error; err != nil {
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

		thumbDimensions, err := image_helpers.GetPhotoDimensions(thumbImagePath)
		if err != nil {
			return false, errors.Wrap(err, "get dimensions of video thumbnail image")
		}

		fileStats, err := os.Stat(thumbImagePath)
		if err != nil {
			return false, errors.Wrap(err, "reading file stats of video thumbnail")
		}

		thumbMediaURL := models.MediaURL{
			MediaID:     video.ID,
			MediaName:   video_thumb_name,
			Width:       thumbDimensions.Width,
			Height:      thumbDimensions.Height,
			Purpose:     models.VideoThumbnail,
			ContentType: "image/jpeg",
			FileSize:    fileStats.Size(),
		}

		if err := tx.Create(&thumbMediaURL).Error; err != nil {
			return false, errors.Wrapf(err, "failed to insert video thumbnail image into database (%s)", video.Title)
		}
	} else {
		// Verify that video thumbnail still exists in cache
		thumbImagePath := path.Join(*videoCachePath, videoThumbnailURL.MediaName)

		if _, err := os.Stat(thumbImagePath); os.IsNotExist(err) {
			fmt.Printf("Video thumbnail found in database but not in cache, re-encoding photo to cache: %s\n", videoThumbnailURL.MediaName)
			didProcess = true

			err = FfmpegCli.EncodeVideoThumbnail(video.Path, thumbImagePath, mediaData)
			if err != nil {
				return false, errors.Wrapf(err, "failed to generate thumbnail for video (%s)", video.Title)
			}

			thumbDimensions, err := image_helpers.GetPhotoDimensions(thumbImagePath)
			if err != nil {
				return false, errors.Wrap(err, "get dimensions of video thumbnail image")
			}

			fileStats, err := os.Stat(thumbImagePath)
			if err != nil {
				return false, errors.Wrap(err, "reading file stats of video thumbnail")
			}

			videoThumbnailURL.Width = thumbDimensions.Width
			videoThumbnailURL.Height = thumbDimensions.Height
			videoThumbnailURL.FileSize = fileStats.Size()

			if err := tx.Save(videoThumbnailURL).Error; err != nil {
				return false, errors.Wrap(err, "updating video thumbnail url in database after re-encoding")
			}
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
