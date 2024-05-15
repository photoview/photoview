package scanner_tasks

import (
	"log"

	"github.com/kkovaletp/photoview/api/graphql/models"
	"github.com/kkovaletp/photoview/api/scanner/exif"
	"github.com/kkovaletp/photoview/api/scanner/scanner_task"
)

type ExifTask struct {
	scanner_task.ScannerTaskBase
}

func (t ExifTask) AfterMediaFound(ctx scanner_task.TaskContext, media *models.Media, newMedia bool) error {

	if !newMedia {
		return nil
	}

	_, err := exif.SaveEXIF(ctx.GetDB(), media)
	if err != nil {
		log.Printf("WARN: SaveEXIF for %s failed: %s\n", media.Title, err)
	}

	return nil
}
