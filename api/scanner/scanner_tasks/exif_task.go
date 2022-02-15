package scanner_tasks

import (
	"log"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/exif"
	"github.com/photoview/photoview/api/scanner/media_encoding"
	"github.com/photoview/photoview/api/scanner/scanner_task"
)

type ExifTask struct {
	scanner_task.ScannerTaskBase
}

func (t ExifTask) AfterProcessMedia(ctx scanner_task.TaskContext, mediaData *media_encoding.EncodeMediaData, updatedURLs []*models.MediaURL, mediaIndex int, mediaTotal int) error {

	_, err := exif.SaveEXIF(ctx.GetDB(), mediaData.Media)
	if err != nil {
		log.Printf("WARN: SaveEXIF for %s failed: %s\n", mediaData.Media.Title, err)
	}

	return nil
}
