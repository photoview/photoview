package migrations

import (
	"fmt"

	"github.com/photoview/photoview/api/graphql/models"
	"gorm.io/gorm"
)

// MigrateCanUploadToAlbumLevel translates the removed, global User.CanUpload
// boolean into a per-album permission level: every existing user_albums row
// for a CanUpload=true user becomes AlbumPermissionLevelDelete (they could
// previously create/upload/move/delete anywhere they owned), and every row
// for a CanUpload=false user becomes AlbumPermissionLevelRead (they were
// view-only) - preserving each user's exact current effective access. Must
// run after AutoMigrate has added user_albums.level, and before the
// can_upload column is dropped.
func MigrateCanUploadToAlbumLevel(db *gorm.DB) error {
	if !db.Migrator().HasColumn(&models.User{}, "can_upload") {
		// Already migrated (or a fresh install with nothing to migrate).
		return nil
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(
			"UPDATE user_albums SET level = ? WHERE user_id IN (SELECT id FROM users WHERE can_upload = true)",
			models.AlbumPermissionLevelDelete,
		).Error; err != nil {
			return fmt.Errorf("failed to migrate can_upload=true users to Delete level: %w", err)
		}

		if err := tx.Exec(
			"UPDATE user_albums SET level = ? WHERE user_id IN (SELECT id FROM users WHERE can_upload = false)",
			models.AlbumPermissionLevelRead,
		).Error; err != nil {
			return fmt.Errorf("failed to migrate can_upload=false users to Read level: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	// DDL (dropping a column) is kept outside the data-migration transaction,
	// matching how the other schema cleanups in MigrateDatabase drop columns
	// directly on db rather than inside a transaction.
	if err := db.Migrator().DropColumn(&models.User{}, "can_upload"); err != nil {
		return fmt.Errorf("failed to drop can_upload column: %w", err)
	}

	return nil
}
