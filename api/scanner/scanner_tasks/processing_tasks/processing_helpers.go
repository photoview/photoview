package processing_tasks

import (
	"fmt"
	"os"
	"path"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/media_encoding"
	"github.com/photoview/photoview/api/scanner/media_encoding/media_utils"
	"github.com/photoview/photoview/api/utils"
	"gorm.io/gorm"
)

// Higher order function used to check if MediaURL for a given MediaPurpose exists
func makePhotoURLChecker(tx *gorm.DB, mediaID int) func(purpose models.MediaPurpose) (*models.MediaURL, error) {
	return func(purpose models.MediaPurpose) (*models.MediaURL, error) {
		var mediaURL []*models.MediaURL

		result := tx.Where("purpose = ?", purpose).Where("media_id = ?", mediaID).Find(&mediaURL)

		if result.Error != nil {
			return nil, result.Error
		}

		if result.RowsAffected > 0 {
			return mediaURL[0], nil
		}

		return nil, nil
	}
}

func generateUniqueMediaNamePrefixed(prefix string, mediaPath string, extension string) string {
	mediaName := fmt.Sprintf("%s_%s_%s", prefix, path.Base(mediaPath), utils.GenerateToken())
	mediaName = models.SanitizeMediaName(mediaName)
	mediaName = mediaName + extension
	return mediaName
}

func generateUniqueMediaName(mediaPath string) string {

	filename := path.Base(mediaPath)
	baseName := filename[0 : len(filename)-len(path.Ext(filename))]
	baseExt := path.Ext(filename)

	mediaName := fmt.Sprintf("%s_%s", baseName, utils.GenerateToken())
	mediaName = models.SanitizeMediaName(mediaName) + baseExt

	return mediaName
}

func saveOriginalPhotoToDB(tx *gorm.DB, photo *models.Media, imageData *media_encoding.EncodeMediaData, photoDimensions *media_utils.PhotoDimensions) (*models.MediaURL, error) {
	originalImageName := generateUniqueMediaName(photo.Path)

	contentType, err := imageData.ContentType()
	if err != nil {
		return nil, err
	}

	fileStats, err := os.Stat(photo.Path)
	if err != nil {
		return nil, fmt.Errorf("reading file stats of original photo: %w", err)
	}

	mediaURL := models.MediaURL{
		Media:       photo,
		MediaName:   originalImageName,
		Width:       photoDimensions.Width,
		Height:      photoDimensions.Height,
		Purpose:     models.MediaOriginal,
		ContentType: string(*contentType),
		FileSize:    fileStats.Size(),
	}

	if err := tx.Create(&mediaURL).Error; err != nil {
		return nil, fmt.Errorf("inserting original photo url: (%d, %s): %w", photo.ID, photo.Title, err)
	}

	return &mediaURL, nil
}
