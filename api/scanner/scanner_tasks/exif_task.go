package scanner_tasks

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/log"
	"github.com/photoview/photoview/api/scanner/externaltools/exif"
	"github.com/photoview/photoview/api/scanner/scanner_task"
)

type ExifTask struct {
	scanner_task.ScannerTaskBase
}

func (t ExifTask) AfterMediaFound(ctx scanner_task.TaskContext, media *models.Media, newMedia bool) error {
	if !newMedia {
		return nil
	}

	if err := SaveEXIF(ctx.GetDB(), media); err != nil {
		log.Warn(ctx, "SaveEXIF failed", "title", media.Title, "error", err, "path", media.Path)
	}

	return nil
}

// SaveEXIF scans the media file for exif metadata and saves it in the database if found
func SaveEXIF(tx *gorm.DB, media *models.Media) error {
	// Check if EXIF data already exists
	if media.ExifID != nil {
		var e models.MediaEXIF
		var err error
		if err = tx.First(&e, media.ExifID).Error; err == nil {
			return nil
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to get EXIF for %q from database: %w", media.Path, err)
		}
		log.Warn(
			tx.Statement.Context,
			"EXIF metadata not found in database, will re-parse it",
			"path", media.Path,
			"error", err,
		)
		media.ExifID = nil
	}

	exifData, err := exif.Parse(media.Path)
	if err != nil {
		return fmt.Errorf("failed to parse exif data: %w", err)
	}

	if exifData == nil {
		return nil
	}

	// Add EXIF to database and link to media
	if err := tx.Model(media).Association("Exif").Replace(exifData); err != nil {
		return fmt.Errorf("failed to save media exif to database: %w", err)
	}

	if exifData.DateShot != nil && !exifData.DateShot.Equal(media.DateShot) {
		media.DateShot = *exifData.DateShot
		if err := tx.Save(media).Error; err != nil {
			return fmt.Errorf("failed to update EXIF metadata for the media %s: %w", media.Path, err)
		}
	}

	return nil
}
