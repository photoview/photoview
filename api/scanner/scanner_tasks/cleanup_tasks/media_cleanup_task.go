package cleanup_tasks

import (
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/scanner_task"
	"github.com/photoview/photoview/api/scanner/scanner_utils"
	"github.com/photoview/photoview/api/scanner/scanner_utils/downloader"
)

type MediaCleanupTask struct {
	scanner_task.ScannerTaskBase
}

func (t MediaCleanupTask) AfterScanAlbum(
	ctx scanner_task.TaskContext,
	changedMedia []*models.Media,
	albumMedia []*models.Media,
) error {
	albumID := ctx.GetAlbum().ID

	cleanupErrors := CleanupMedia(ctx.GetDB(), ctx.GetCacheFS(), albumID, albumMedia)
	for _, err := range cleanupErrors {
		scanner_utils.ScannerError(ctx, "delete old media: %s", err)
	}

	// Delete temporary files used during scanning
	err := downloader.CleanupTempFiles(albumID)
	if err != nil {
		scanner_utils.ScannerError(ctx, "cleanup temp files: %s", err)
	}

	return nil
}
