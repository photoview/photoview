package migrations

import (
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// MigrateForExifGPSCorrection finds and removes invalid GPS data from media_exif table
func MigrateForExifGPSCorrection(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.MediaEXIF{}).
			Where("ABS(gps_longitude) > ?", 90).
			Or("ABS(gps_latitude) > ?", 90).
			Updates(map[string]interface{}{
				"gps_latitude":  nil,
				"gps_longitude": nil,
			}).Error; err != nil {
			return errors.Wrap(err, "failed to remove invalid GPS data from media_exif table")
		}
		return nil
	})
}
