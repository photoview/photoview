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

	// Re-encode to h264 if the web video cache is missing.
	var videoWebURLStatErr error
	// to prevent nil pointer dereferece
	if videoWebURL != nil && len(videoWebURL.MediaName) > 0 {
		_, videoWebURLStatErr = os.Stat(path.Join(mediaCachePath, videoWebURL.MediaName))
	}

	if videoOriginalURL == nil && (videoWebURL == nil || videoWebURLStatErr != nil) {
		// Decide whether the original video file is compatible with web browsers.
		origVideoPath := video.Path
		videoMediaName := generateUniqueMediaName(origVideoPath)

		// Read video and audio codec names.
		origFileMetadata, err := ReadVideoMetadata(origVideoPath)
		if err != nil {
			return nil, errors.Wrap(err, "read video stream metadata")
		}
		origVideoStream := origFileMetadata.FirstVideoStream()
		if origVideoStream == nil {
			return []*models.MediaURL{}, errors.Wrapf(err, "could not get video stream from file metadata (%s)", origVideoPath)
		}
		origAudioStream := origFileMetadata.FirstAudioStream()

		if videoType.IsWebCompatible() && videoCodecIsWebCompatible(origVideoStream) && audioCodecIsWebCompatible(origAudioStream) {
			log.Printf("Video has a compatible container and : %s", video.Path)

			fileStats, err := os.Stat(origVideoPath)
			if err != nil {
				return []*models.MediaURL{}, errors.Wrap(err, "reading file stats of original video")
			}

			mediaURL := models.MediaURL{
				MediaID:     video.ID,
				MediaName:   videoMediaName,
				Width:       origVideoStream.Width,
				Height:      origVideoStream.Height,
				Purpose:     models.MediaOriginal,
				ContentType: string(*videoType),
				FileSize:    fileStats.Size(),
			}

			if err := ctx.GetDB().Create(&mediaURL).Error; err != nil {
				return []*models.MediaURL{}, errors.Wrapf(err, "insert original video into database (%s)", video.Title)
			}

			updatedURLs = append(updatedURLs, &mediaURL)
		} else {
			web_video_name := fmt.Sprintf("web_video_%s_%s", path.Base(video.Path), utils.GenerateToken())
			web_video_name = strings.ReplaceAll(web_video_name, ".", "_")
			web_video_name = strings.ReplaceAll(web_video_name, " ", "_")
			web_video_name = web_video_name + ".mp4"

			webVideoPath := path.Join(mediaCachePath, web_video_name)

			err = executable_worker.FfmpegCli.EncodeMp4(video.Path, webVideoPath)
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
				MediaName:   web_video_name,
				Width:       webMetadata.Width,
				Height:      webMetadata.Height,
				Purpose:     models.VideoWeb,
				ContentType: "video/mp4",
				FileSize:    fileStats.Size(),
			}

			if videoWebURL == nil { // Newly encoded video.
				if err := ctx.GetDB().Create(&mediaURL).Error; err != nil {
					return []*models.MediaURL{}, errors.Wrapf(err, "failed to insert encoded web-video into database (%s)", video.Title)
				}
			} else { // A missing video cache is restored.
				if err := ctx.GetDB().Save(&mediaURL).Error; err != nil {
					return []*models.MediaURL{}, errors.Wrapf(err, "failed to insert encoded web-video into database (%s)", video.Title)
				}
			}

			updatedURLs = append(updatedURLs, &mediaURL)
		}
	}

	probeData, err := mediaData.VideoMetadata()
	if err != nil {
		return []*models.MediaURL{}, err
	}

	if videoThumbnailURL == nil {
		video_thumb_name := fmt.Sprintf("video_thumb_%s_%s", path.Base(video.Path), utils.GenerateToken())
		video_thumb_name = strings.ReplaceAll(video_thumb_name, ".", "_")
		video_thumb_name = strings.ReplaceAll(video_thumb_name, " ", "_")
		video_thumb_name = video_thumb_name + ".jpg"

		thumbImagePath := path.Join(mediaCachePath, video_thumb_name)

		err = executable_worker.FfmpegCli.EncodeVideoThumbnail(video.Path, thumbImagePath, probeData)
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
			MediaName:   video_thumb_name,
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

			err = executable_worker.FfmpegCli.EncodeVideoThumbnail(video.Path, thumbImagePath, probeData)
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

// audioCodecIsWebCompatible returns true if the audio codec is compatible with
// web browsers, as defined by https://en.wikipedia.org/wiki/HTML5_audio#Supported_audio_coding_formats
func audioCodecIsWebCompatible(audio *ffprobe.Stream) bool {
	if audio == nil ||
		audio.CodecName == "aac" ||
		audio.CodecName == "mp3" ||
		audio.CodecName == "opus" ||
		audio.CodecName == "flac" ||
		audio.CodecName == "vorbis" {
		return true
	}
	return false
}

// videoCodecIsWebCompatible returns true if the video codec is compatible with
// web browsers, as defined by https://en.wikipedia.org/wiki/HTML5_video#Browser_support
func videoCodecIsWebCompatible(video *ffprobe.Stream) bool {
	if video.CodecName == "h264" || video.CodecName == "vp8" || video.CodecName == "vp9" || video.CodecName == "av1" {
		return true
	}
	return false
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
		return nil, errors.Wrapf(err, "could not get stream from file metadata (%s)", path.Base(videoPath))
	}

	return stream, nil
}
