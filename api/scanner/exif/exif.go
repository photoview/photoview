package exif

import (
	"sync"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/log"
)

type ExifParser interface {
	ParseExif(media_path string) (*models.MediaEXIF, error)
}

var globalExifParser ExifParser
var globalInit sync.Once

func InitializeEXIFParser() {
	globalInit.Do(func() {
		// Decide between internal or external Exif parser
		exiftoolParser, err := NewExiftoolParser()
		if err != nil {
			log.Warn(nil, "Failed to get exiftool, using internal exif parser instead", "exif_error", err)
			globalExifParser = NewInternalExifParser()
			return
		}

		log.Info(nil, "Found exiftool")
		globalExifParser = exiftoolParser
	})
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

	if globalExifParser == nil {
		return nil, errors.New("No exif parser initialized")
	}

	exif, err := globalExifParser.ParseExif(media.Path)
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
