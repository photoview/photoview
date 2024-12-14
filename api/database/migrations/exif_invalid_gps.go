package migrations

import (
	"fmt"

	"github.com/photoview/photoview/api/graphql/models"
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
			return fmt.Errorf("failed to remove invalid GPS data from media_exif table: %w", err)
		}
		return nil
	})
}
