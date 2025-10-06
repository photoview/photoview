package scanner_tasks

import (
	"fmt"
	"time"

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
		var exif models.MediaEXIF
		if err := tx.First(&exif, media.ExifID).Error; err != nil {
			return fmt.Errorf("failed to get EXIF for %q from database: %w", media.Path, err)
		}

		return nil
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

	if exifData.DateShotStr == nil {
		return nil
	}

	dateShot, err := time.Parse(models.RFC3339Milli, *exifData.DateShotStr)
	if err != nil {
		dateShot, err = time.ParseInLocation(models.RFC3339MilliWithoutTimezone, *exifData.DateShotStr, time.Local)
		if err != nil {
			return fmt.Errorf("invalid dateshot when parsing exif data: %w", err)
		}
	}

	media.DateShot = dateShot
	if err := tx.Save(media).Error; err != nil {
		return fmt.Errorf("failed to update media date_shot: %w", err)
	}

	return nil
}
