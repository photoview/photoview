package migrations

import (
	"fmt"

	"github.com/photoview/photoview/api/graphql/models"
	"gorm.io/gorm"
)

// MigrateForDuplicateMediaURLs removes duplicate media_urls rows sharing the
// same (media_id, purpose) pair, keeping only the most recently updated one.
// This must run before AutoMigrate adds a unique index on that pair, so that
// an install that somehow already accumulated duplicates doesn't fail to
// migrate.
func MigrateForDuplicateMediaURLs(db *gorm.DB) error {
	if !db.Migrator().HasTable(&models.MediaURL{}) {
		// Fresh database: AutoMigrate will create the table (with the new
		// unique index already in place) from scratch, nothing to dedup yet.
		return nil
	}

	return db.Transaction(func(tx *gorm.DB) error {
		var urls []models.MediaURL
		if err := tx.Order("media_id, purpose, updated_at DESC, id DESC").Find(&urls).Error; err != nil {
			return fmt.Errorf("find media urls: %w", err)
		}

		seen := make(map[string]bool, len(urls))
		var staleIDs []int
		for _, u := range urls {
			key := fmt.Sprintf("%d:%s", u.MediaID, u.Purpose)
			if seen[key] {
				staleIDs = append(staleIDs, u.ID)
				continue
			}
			seen[key] = true
		}

		if len(staleIDs) == 0 {
			return nil
		}

		if err := tx.Delete(&models.MediaURL{}, staleIDs).Error; err != nil {
			return fmt.Errorf("delete duplicate media urls: %w", err)
		}

		return nil
	})
}
