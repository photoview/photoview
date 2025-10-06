package migrations

import (
	"fmt"
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"gorm.io/gorm"
)

func MigrateDateShot(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		rows, err := tx.Model(&models.MediaEXIF{}).Select("date_shot").Where("date_shot IS NOT NULL AND date_shot_str IS NULL").Rows()
		if err != nil {
			return fmt.Errorf("can't query rows for migrating date_shot: %w", err)
		}

		for rows.Next() {
			var dateShot time.Time
			if err := rows.Scan(&dateShot); err != nil {
				return fmt.Errorf("can't query rows for date_shot: %w", err)
			}

			if err := tx.Model(&models.MediaEXIF{}).Update("date_shot_str", dateShot.Format(models.RFC3339MilliWithoutTimezone)).Error; err != nil {
				return fmt.Errorf("can't update rows for date_shot_at: %w", err)
			}
		}

		return nil
	})
}
