package exif

import (
	"log"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/photoview/photoview/api/graphql/models"
)

type exifParser interface {
	ParseExif(media *models.Media) (*models.MediaEXIF, error)
}

// SaveEXIF scans the media file for exif metadata and saves it in the database if found
func SaveEXIF(tx *gorm.DB, media *models.Media) (*models.MediaEXIF, error) {

	log.Printf("Scanning for EXIF: %s", media.Path)

	{
		// Check if EXIF data already exists
		if media.ExifID != nil {

			var exif models.MediaEXIF
			if err := tx.First(&exif, media.ExifID).Error; err != nil {
				return nil, errors.Wrap(err, "get EXIF for media from database")
			}

			return &exif, nil
		}
	}

	var parser exifParser = &internalExifParser{}

	exif, err := parser.ParseExif(media)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse exif data")
	}

	if exif == nil {
		return nil, nil
	}

	// Add EXIF to database and link to media
	if err := tx.Model(&media).Association("Exif").Replace(exif); err != nil {
		return nil, errors.Wrap(err, "save media exif to database")
	}

	return exif, nil
}
