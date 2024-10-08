package exif

import (
	"fmt"
	"log"

	"gorm.io/gorm"

	"github.com/photoview/photoview/api/graphql/models"
)

type ExifParser interface {
	ParseExif(media_path string) (*models.MediaEXIF, error)
}

var globalExifParser ExifParser

func InitializeEXIFParser() {
	// Decide between internal or external Exif parser
	exiftoolParser, err := NewExiftoolParser()

	if err != nil {
		log.Printf("Failed to get exiftool, using internal exif parser instead: %v\n", err)
		globalExifParser = NewInternalExifParser()
	} else {
		log.Println("Found exiftool")
		globalExifParser = exiftoolParser
	}
}

// SaveEXIF scans the media file for exif metadata and saves it in the database if found
func SaveEXIF(tx *gorm.DB, media *models.Media) (*models.MediaEXIF, error) {

	{
		// Check if EXIF data already exists
		if media.ExifID != nil {

			var exif models.MediaEXIF
			if err := tx.First(&exif, media.ExifID).Error; err != nil {
				return nil, fmt.Errorf("get EXIF for media from database: %w", err)
			}

			return &exif, nil
		}
	}

	if globalExifParser == nil {
		return nil, fmt.Errorf("no exif parser initialized")
	}

	exif, err := globalExifParser.ParseExif(media.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse exif data: %w", err)
	}

	if exif == nil {
		return nil, nil
	}

	// Add EXIF to database and link to media
	if err := tx.Model(&media).Association("Exif").Replace(exif); err != nil {
		return nil, fmt.Errorf("save media exif to database: %w", err)
	}

	if exif.DateShot != nil && !exif.DateShot.Equal(media.DateShot) {
		media.DateShot = *exif.DateShot
		if err := tx.Save(media).Error; err != nil {
			return nil, fmt.Errorf("update media date_shot: %w", err)
		}
	}

	return exif, nil
}
