package scanner_tasks

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"os"

	"github.com/buckket/go-blurhash"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/log"
	"github.com/photoview/photoview/api/scanner/media_encoding"
	"github.com/photoview/photoview/api/scanner/scanner_task"
)

type BlurhashTask struct {
	scanner_task.ScannerTaskBase
}

func (t BlurhashTask) AfterProcessMedia(ctx scanner_task.TaskContext, mediaData *media_encoding.EncodeMediaData, updatedURLs []*models.MediaURL, mediaIndex int, mediaTotal int) error {
	hasThumbnailUpdated := false
	for _, url := range updatedURLs {
		if url.Purpose == models.PhotoThumbnail || url.Purpose == models.VideoThumbnail {
			hasThumbnailUpdated = true
			break
		}
	}

	media := mediaData.Media
	if media.Blurhash != nil && !hasThumbnailUpdated {
		return nil
	}

	thumbnail, err := media.GetThumbnail()
	if err != nil {
		return fmt.Errorf("failed to get thubmnail of image %q: %w", mediaData.Media.Path, err)
	}

	hashStr, err := generateBlurhashFromThumbnail(thumbnail)
	if err != nil {
		return fmt.Errorf("failed to generate blurhash of image %q: %w", mediaData.Media.Path, err)
	}

	media.Blurhash = &hashStr
	if err := ctx.GetDB().Select("blurhash").Save(media).Error; err != nil {
		return fmt.Errorf("failed to store blurhash of image %q: %w", mediaData.Media.Path, err)
	}

	log.Info(ctx, "Generated blurhash of image %q", mediaData.Media.Path)

	return nil
}

// generateBlurhashFromThumbnail generates a blurhash for a single media and stores it in the database
func generateBlurhashFromThumbnail(thumbnail *models.MediaURL) (string, error) {
	thumbnail_path, err := thumbnail.CachedPath()
	if err != nil {
		return "", fmt.Errorf("get path of media(id:%d) error: %w", thumbnail.MediaID, err)
	}

	imageFile, err := os.Open(thumbnail_path)
	if err != nil {
		return "", fmt.Errorf("open %q error: %w", thumbnail_path, err)
	}
	defer imageFile.Close()

	imageData, _, err := image.Decode(imageFile)
	if err != nil {
		return "", fmt.Errorf("decode %q error: %w", thumbnail_path, err)
	}

	hashStr, err := blurhash.Encode(4, 3, imageData)
	if err != nil {
		return "", fmt.Errorf("encode blurhash of %q error: %w", thumbnail_path, err)
	}

	return hashStr, nil
}
