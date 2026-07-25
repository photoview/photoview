// Package migration_exif contains the migration for EXIF data.
package migration_exif

import (
	"context"
	"fmt"

	"github.com/photoview/photoview/api/database"
	"github.com/photoview/photoview/api/graphql/models"
)

// MigrateExif migrates the EXIF data.
func MigrateExif(ctx context.Context) error {
	// Migrate the EXIF data.
	return database.MigrateExif(ctx)
}