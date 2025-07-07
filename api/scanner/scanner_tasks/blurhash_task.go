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

	var media *models.Media
	if err := ctx.GetDB().Preload("MediaURL").Where("id = ?", mediaData.Media.ID).First(&media).Error; err != nil {
		return fmt.Errorf("failed to get media(id:%d): %w", mediaData.Media.ID, err)
	}

	if media.Blurhash != nil && !hasThumbnailUpdated {
		log.Info(ctx, "No thumbnail updated, ignore generating blurhash", "media", media.Path)
		return nil
	}

	thumbnail, err := media.GetThumbnail()
	if err != nil {
		return fmt.Errorf("failed to get thumbnail of image %q: %w", mediaData.Media.Path, err)
	}

	hashStr, err := generateBlurhashFromThumbnail(thumbnail)
	if err != nil {
		return fmt.Errorf("failed to generate blurhash of image %q: %w", mediaData.Media.Path, err)
	}

	media.Blurhash = &hashStr
	if err := ctx.GetDB().Select("blurhash").Save(media).Error; err != nil {
		return fmt.Errorf("failed to store blurhash of image %q: %w", mediaData.Media.Path, err)
	}

	log.Info(ctx, "Generated blurhash of image", "media", mediaData.Media.Path)

	return nil
}

// generateBlurhashFromThumbnail generates a blurhash for a single media and stores it in the database
func generateBlurhashFromThumbnail(thumbnail *models.MediaURL) (string, error) {
	path, err := thumbnail.CachedPath()
	if err != nil {
		return "", fmt.Errorf("get path of media(id:%d) error: %w", thumbnail.MediaID, err)
	}

	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open %q error: %w", path, err)
	}
	defer f.Close()

	imageData, _, err := image.Decode(f)
	if err != nil {
		return "", fmt.Errorf("decode %q error: %w", path, err)
	}

	hashStr, err := blurhash.Encode(4, 3, imageData)
	if err != nil {
		return "", fmt.Errorf("encode blurhash of %q error: %w", path, err)
	}

	return hashStr, nil
}
