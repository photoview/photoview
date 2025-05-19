package scanner

import (
	"fmt"
	"image"
	"log"
	"os"

	"github.com/buckket/go-blurhash"
	"github.com/photoview/photoview/api/graphql/models"
	"gorm.io/gorm"
)

// GenerateBlurhashes queries the database for media that are missing a blurhash and computes one for them.
// This function blocks until all hashes have been computed
func GenerateBlurhashes(db *gorm.DB) error {
	var results []*models.Media

	processErrors := make([]error, 0)

	query := db.Model(&models.Media{}).
		Preload("MediaURL").
		Joins("INNER JOIN media_urls ON media.id = media_urls.media_id").
		Where("blurhash IS NULL").
		Where("media_urls.purpose = 'thumbnail' OR media_urls.purpose = 'video-thumbnail'")

	err := query.FindInBatches(&results, 50, func(tx *gorm.DB, batch int) error {
		log.Printf("generating %d blurhashes", len(results))

		hashes := make([]*string, len(results))

		for i, row := range results {

			thumbnail, err := row.GetThumbnail()
			if err != nil {
				log.Printf("failed to get thumbnail for media to generate blurhash (%d): %v", row.ID, err)
				processErrors = append(processErrors, err)
				continue
			}

			hashStr, err := GenerateBlurhashFromThumbnail(thumbnail)
			if err != nil {
				log.Printf("failed to generate blurhash: %v", err)
				processErrors = append(processErrors, err)
				continue
			}

			hashes[i] = &hashStr
			results[i].Blurhash = &hashStr
		}

		tx.Save(results)
		// if err := db.Update("blurhash", hashes).Error; err != nil {
		// 	return err
		// }

		return nil
	}).Error

	if err != nil {
		return err
	}

	if len(processErrors) == 0 {
		return nil
	} else {
		return fmt.Errorf("failed to generate %d blurhashes", len(processErrors))
	}
}

// GenerateBlurhashFromThumbnail generates a blurhash for a single media and stores it in the database
func GenerateBlurhashFromThumbnail(thumbnail *models.MediaURL) (string, error) {
	thumbnail_path, err := thumbnail.CachedPath()
	if err != nil {
		return "", fmt.Errorf("get path of media id=%d error: %w", thumbnail.MediaID, err)
	}

	imageFile, err := os.Open(thumbnail_path)
	if err != nil {
		return "", fmt.Errorf("open %s error: %w", thumbnail_path, err)
	}

	imageData, _, err := image.Decode(imageFile)
	if err != nil {
		return "", fmt.Errorf("decode %q error: %w", thumbnail_path, err)
	}

	hashStr, err := blurhash.Encode(4, 3, imageData)
	if err != nil {
		return "", fmt.Errorf("encode blurhash of %q error: %w", thumbnail_path, err)
	}

	// if err := db.Model(&models.Media{}).Where("id = ?", thumbnail.MediaID).Update("blurhash", hashStr).Error; err != nil {
	// return "", fmt.Errorf("update blurhash of media id=%d error: %w", thumbnail.MediaID, err)
	// }

	return hashStr, nil
}
