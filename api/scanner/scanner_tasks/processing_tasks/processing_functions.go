package processing_tasks

import (
	"fmt"
	"os"
	"path"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/media_encoding"
	"github.com/photoview/photoview/api/scanner/media_encoding/media_utils"
	"gorm.io/gorm"
)

func generateSaveHighResJPEG(tx *gorm.DB, media *models.Media, imageData *media_encoding.EncodeMediaData, highResName string, imagePath string, mediaURL *models.MediaURL) (*models.MediaURL, error) {

	err := imageData.EncodeHighRes(imagePath)
	if err != nil {
		return nil, fmt.Errorf("creating high-res cached image: %w", err)
	}

	photoDimensions, err := media_utils.GetPhotoDimensions(imagePath)
	if err != nil {
		return nil, err
	}

	fileStats, err := os.Stat(imagePath)
	if err != nil {
		return nil, fmt.Errorf("reading file stats of highres photo: %w", err)
	}

	if mediaURL == nil {

		mediaURL = &models.MediaURL{
			MediaID:     media.ID,
			MediaName:   highResName,
			Width:       photoDimensions.Width,
			Height:      photoDimensions.Height,
			Purpose:     models.PhotoHighRes,
			ContentType: "image/jpeg",
			FileSize:    fileStats.Size(),
		}

		if err := tx.Create(&mediaURL).Error; err != nil {
			return nil, fmt.Errorf("could not insert highres media url (%d, %s): %w", media.ID, highResName, err)
		}
	} else {
		mediaURL.Width = photoDimensions.Width
		mediaURL.Height = photoDimensions.Height
		mediaURL.FileSize = fileStats.Size()

		if err := tx.Save(&mediaURL).Error; err != nil {
			return nil,
				fmt.Errorf("could not update media url after side car changes (%d, %s): %w", media.ID, highResName, err)
		}
	}

	return mediaURL, nil
}

func generateSaveThumbnailJPEG(tx *gorm.DB, media *models.Media, thumbnailName string, photoCachePath string, baseImagePath string, mediaURL *models.MediaURL) (*models.MediaURL, error) {
	thumbOutputPath := path.Join(photoCachePath, thumbnailName)

	thumbSize, err := media_encoding.EncodeThumbnail(tx, baseImagePath, thumbOutputPath)
	if err != nil {
		return nil, fmt.Errorf("could not create thumbnail cached image: %w", err)
	}

	fileStats, err := os.Stat(thumbOutputPath)
	if err != nil {
		return nil, fmt.Errorf("reading file stats of thumbnail photo: %w", err)
	}

	if mediaURL == nil {

		mediaURL = &models.MediaURL{
			MediaID:     media.ID,
			MediaName:   thumbnailName,
			Width:       thumbSize.Width,
			Height:      thumbSize.Height,
			Purpose:     models.PhotoThumbnail,
			ContentType: "image/jpeg",
			FileSize:    fileStats.Size(),
		}

		if err := tx.Create(&mediaURL).Error; err != nil {
			return nil, fmt.Errorf("could not insert thumbnail media url (%d, %s): %w", media.ID, thumbnailName, err)
		}
	} else {
		mediaURL.Width = thumbSize.Width
		mediaURL.Height = thumbSize.Height
		mediaURL.FileSize = fileStats.Size()

		if err := tx.Save(&mediaURL).Error; err != nil {
			return nil,
				fmt.Errorf("could not update media url after side car changes (%d, %s): %w", media.ID, thumbnailName, err)
		}
	}

	return mediaURL, nil
}
