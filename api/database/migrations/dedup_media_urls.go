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
		// A row is stale if some other row shares its (media_id, purpose)
		// pair and outranks it (higher updated_at, ties broken by higher
		// id) - the same tie-break rule the old in-memory version used.
		// The "stale" derived table works around MySQL's restriction on
		// referencing the delete target directly in a subquery's FROM.
		if err := tx.Exec(`
			DELETE FROM media_urls
			WHERE id IN (
				SELECT id FROM (
					SELECT older.id
					FROM media_urls AS older
					JOIN media_urls AS newer
						ON newer.media_id = older.media_id
						AND newer.purpose = older.purpose
						AND (
							newer.updated_at > older.updated_at
							OR (newer.updated_at = older.updated_at AND newer.id > older.id)
						)
				) AS stale
			)
		`).Error; err != nil {
			return fmt.Errorf("delete duplicate media urls: %w", err)
		}

		return nil
	})
}
