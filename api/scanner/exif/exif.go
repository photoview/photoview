package exif

import (
	"log"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/barasher/go-exiftool"
	"github.com/photoview/photoview/api/graphql/models"
)

type exifParser interface {
	ParseExif(media_path string) (*models.MediaEXIF, error)
}

var use_exiftool bool = false

func InitializeEXIFParser() {
	// Decide between internal or external Exif parser
	et, err := exiftool.NewExiftool()

	if err != nil {
		use_exiftool = false
		log.Printf("Failed to get exiftool, using internal exif parser instead: %v\n", err)
	} else {
		et.Close()
		log.Println("Found exiftool")
		use_exiftool = true
	}
}

// SaveEXIF scans the media file for exif metadata and saves it in the database if found
func SaveEXIF(tx *gorm.DB, media *models.Media) (*models.MediaEXIF, error) {

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

	var parser exifParser
	if use_exiftool {
		parser = &externalExifParser{}
	} else {
		parser = &internalExifParser{}
	}

	exif, err := parser.ParseExif(media.Path)
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

	if exif.DateShot != nil && !exif.DateShot.Equal(media.DateShot) {
		media.DateShot = *exif.DateShot
		if err := tx.Save(media).Error; err != nil {
			return nil, errors.Wrap(err, "update media date_shot")
		}
	}

	return exif, nil
}
