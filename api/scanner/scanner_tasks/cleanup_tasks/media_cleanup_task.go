package cleanup_tasks

import (
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/scanner_task"
	"github.com/photoview/photoview/api/scanner/scanner_utils"
)

type MediaCleanupTask struct {
	scanner_task.ScannerTaskBase
}

func (t MediaCleanupTask) AfterScanAlbum(ctx scanner_task.TaskContext, changedMedia []*models.Media, albumMedia []*models.Media) error {

	cleanupErrors := CleanupMedia(ctx.GetDB(), ctx.GetAlbum().ID, albumMedia)
	for _, err := range cleanupErrors {
		scanner_utils.ScannerError("delete old media: %s", err)
	}

	return nil
}
