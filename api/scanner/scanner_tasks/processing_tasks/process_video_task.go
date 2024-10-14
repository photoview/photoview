package processing_tasks

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/media_encoding"
	"github.com/photoview/photoview/api/scanner/media_encoding/executable_worker"
	"github.com/photoview/photoview/api/scanner/media_encoding/media_utils"
	"github.com/photoview/photoview/api/scanner/scanner_task"
	"github.com/photoview/photoview/api/utils"
	"github.com/pkg/errors"
	"gopkg.in/vansante/go-ffprobe.v2"
)

type ProcessVideoTask struct {
	scanner_task.ScannerTaskBase
}

func (t ProcessVideoTask) ProcessMedia(ctx scanner_task.TaskContext, mediaData *media_encoding.EncodeMediaData, mediaCachePath string) ([]*models.MediaURL, error) {
	if mediaData.Media.Type != models.MediaTypeVideo {
		return []*models.MediaURL{}, nil
	}

	updatedURLs := make([]*models.MediaURL, 0)
	video := mediaData.Media

	log.Printf("Processing video: %s", video.Path)

	mediaURLFromDB := makePhotoURLChecker(ctx.GetDB(), video.ID)

	videoOriginalURL, err := mediaURLFromDB(models.MediaOriginal)
	if err != nil {
		return []*models.MediaURL{}, errors.Wrap(err, "error processing video original format")
	}

	videoWebURL, err := mediaURLFromDB(models.VideoWeb)
	if err != nil {
		return []*models.MediaURL{}, errors.Wrap(err, "error processing video web-format")
	}

	videoThumbnailURL, err := mediaURLFromDB(models.VideoThumbnail)
	if err != nil {
		return []*models.MediaURL{}, errors.Wrap(err, "error processing video thumbnail")
	}

	videoType, err := mediaData.ContentType()
	if err != nil {
		return []*models.MediaURL{}, errors.Wrap(err, "error getting video content type")
	}

	if videoOriginalURL == nil && videoType.IsWebCompatible() {
		origVideoPath := video.Path
		videoMediaName := generateUniqueMediaName(video.Path)

		webMetadata, err := ReadVideoStreamMetadata(origVideoPath)
		if err != nil {
			return []*models.MediaURL{}, errors.Wrapf(err, "failed to read metadata for original video (%s)", video.Title)
		}

		fileStats, err := os.Stat(origVideoPath)
		if err != nil {
			return []*models.MediaURL{}, errors.Wrap(err, "reading file stats of original video")
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

		if err := ctx.GetDB().Create(&mediaURL).Error; err != nil {
			return []*models.MediaURL{}, errors.Wrapf(err, "insert original video into database (%s)", video.Title)
		}

		updatedURLs = append(updatedURLs, &mediaURL)
	}

	if videoWebURL == nil && !videoType.IsWebCompatible() {
		webVideoName := fmt.Sprintf("web_video_%s_%s", path.Base(video.Path), utils.GenerateToken())
		webVideoName = strings.ReplaceAll(webVideoName, ".", "_")
		webVideoName = strings.ReplaceAll(webVideoName, " ", "_")
		webVideoName = webVideoName + ".mp4"

		webVideoPath := path.Join(mediaCachePath, webVideoName)

		err = executable_worker.Ffmpeg.EncodeMp4(video.Path, webVideoPath)
		if err != nil {
			return []*models.MediaURL{}, errors.Wrapf(err, "could not encode mp4 video (%s)", video.Path)
		}

		webMetadata, err := ReadVideoStreamMetadata(webVideoPath)
		if err != nil {
			return []*models.MediaURL{}, errors.Wrapf(err, "failed to read metadata for encoded web-video (%s)", video.Title)
		}

		fileStats, err := os.Stat(webVideoPath)
		if err != nil {
			return []*models.MediaURL{}, errors.Wrap(err, "reading file stats of web-optimized video")
		}

		mediaURL := models.MediaURL{
			MediaID:     video.ID,
			MediaName:   webVideoName,
			Width:       webMetadata.Width,
			Height:      webMetadata.Height,
			Purpose:     models.VideoWeb,
			ContentType: "video/mp4",
			FileSize:    fileStats.Size(),
		}

		if err := ctx.GetDB().Create(&mediaURL).Error; err != nil {
			return []*models.MediaURL{}, errors.Wrapf(err, "failed to insert encoded web-video into database (%s)", video.Title)
		}

		updatedURLs = append(updatedURLs, &mediaURL)
	}

	probeData, err := mediaData.VideoMetadata()
	if err != nil {
		return []*models.MediaURL{}, err
	}

	if videoThumbnailURL == nil {
		videoThumbName := fmt.Sprintf("video_thumb_%s_%s", path.Base(video.Path), utils.GenerateToken())
		videoThumbName = strings.ReplaceAll(videoThumbName, ".", "_")
		videoThumbName = strings.ReplaceAll(videoThumbName, " ", "_")
		videoThumbName = videoThumbName + ".jpg"

		thumbImagePath := path.Join(mediaCachePath, videoThumbName)

		err = executable_worker.Ffmpeg.EncodeVideoThumbnail(video.Path, thumbImagePath, probeData)
		if err != nil {
			return []*models.MediaURL{}, errors.Wrapf(err, "failed to generate thumbnail for video (%s)", video.Title)
		}

		thumbDimensions, err := media_utils.GetPhotoDimensions(thumbImagePath)
		if err != nil {
			return []*models.MediaURL{}, errors.Wrap(err, "get dimensions of video thumbnail image")
		}

		fileStats, err := os.Stat(thumbImagePath)
		if err != nil {
			return []*models.MediaURL{}, errors.Wrap(err, "reading file stats of video thumbnail")
		}

		thumbMediaURL := models.MediaURL{
			MediaID:     video.ID,
			MediaName:   videoThumbName,
			Width:       thumbDimensions.Width,
			Height:      thumbDimensions.Height,
			Purpose:     models.VideoThumbnail,
			ContentType: "image/jpeg",
			FileSize:    fileStats.Size(),
		}

		if err := ctx.GetDB().Create(&thumbMediaURL).Error; err != nil {
			return []*models.MediaURL{}, errors.Wrapf(err, "failed to insert video thumbnail image into database (%s)", video.Title)
		}

		updatedURLs = append(updatedURLs, &thumbMediaURL)
	} else {
		// Verify that video thumbnail still exists in cache
		thumbImagePath := path.Join(mediaCachePath, videoThumbnailURL.MediaName)

		if _, err := os.Stat(thumbImagePath); os.IsNotExist(err) {
			fmt.Printf("Video thumbnail found in database but not in cache, re-encoding photo to cache: %s\n", videoThumbnailURL.MediaName)
			updatedURLs = append(updatedURLs, videoThumbnailURL)

			err = executable_worker.Ffmpeg.EncodeVideoThumbnail(video.Path, thumbImagePath, probeData)
			if err != nil {
				return []*models.MediaURL{}, errors.Wrapf(err, "failed to generate thumbnail for video (%s)", video.Title)
			}

			thumbDimensions, err := media_utils.GetPhotoDimensions(thumbImagePath)
			if err != nil {
				return []*models.MediaURL{}, errors.Wrap(err, "get dimensions of video thumbnail image")
			}

			fileStats, err := os.Stat(thumbImagePath)
			if err != nil {
				return []*models.MediaURL{}, errors.Wrap(err, "reading file stats of video thumbnail")
			}

			videoThumbnailURL.Width = thumbDimensions.Width
			videoThumbnailURL.Height = thumbDimensions.Height
			videoThumbnailURL.FileSize = fileStats.Size()

			if err := ctx.GetDB().Save(videoThumbnailURL).Error; err != nil {
				return []*models.MediaURL{}, errors.Wrap(err, "updating video thumbnail url in database after re-encoding")
			}
		}
	}

	return updatedURLs, nil
}

func ReadVideoMetadata(videoPath string) (*ffprobe.ProbeData, error) {
	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()

	data, err := ffprobe.ProbeURL(ctx, videoPath)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read video metadata (%s)", path.Base(videoPath))
	}

	return data, nil
}

func ReadVideoStreamMetadata(videoPath string) (*ffprobe.Stream, error) {
	data, err := ReadVideoMetadata(videoPath)
	if err != nil {
		return nil, errors.Wrap(err, "read video stream metadata")
	}

	stream := data.FirstVideoStream()
	if stream == nil {
		return nil, fmt.Errorf("could not get stream from file metadata (%s)", path.Base(videoPath))
	}

	return stream, nil
}
