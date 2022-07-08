package scanner_tasks

import (
	"io/fs"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/media_encoding"
	"github.com/photoview/photoview/api/scanner/scanner_task"
	"github.com/photoview/photoview/api/scanner/scanner_tasks/cleanup_tasks"
	"github.com/photoview/photoview/api/scanner/scanner_tasks/processing_tasks"
)

var allTasks []scanner_task.ScannerTask = []scanner_task.ScannerTask{
	NotificationTask{},
	IgnorefileTask{},
	processing_tasks.CounterpartFilesTask{},
	processing_tasks.SidecarTask{},
	processing_tasks.ProcessPhotoTask{},
	processing_tasks.ProcessVideoTask{},
	FaceDetectionTask{},
	ExifTask{},
	VideoMetadataTask{},
	cleanup_tasks.MediaCleanupTask{},
}

type scannerTasks struct {
	scanner_task.ScannerTaskBase
}

var Tasks scannerTasks = scannerTasks{}

func simpleCombinedTasks(ctx scanner_task.TaskContext, doTask func(ctx scanner_task.TaskContext, task scanner_task.ScannerTask) error) error {
	for _, task := range allTasks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := doTask(ctx, task)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t scannerTasks) BeforeScanAlbum(ctx scanner_task.TaskContext) (scanner_task.TaskContext, error) {
	for _, task := range allTasks {
		var err error
		ctx, err = task.BeforeScanAlbum(ctx)
		if err != nil {
			return ctx, err
		}

		select {
		case <-ctx.Done():
			return ctx, ctx.Err()
		default:
		}
	}

	return ctx, nil
}

func (t scannerTasks) MediaFound(ctx scanner_task.TaskContext, fileInfo fs.FileInfo, mediaPath string) (bool, error) {
	for _, task := range allTasks {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
		}

		skip, err := task.MediaFound(ctx, fileInfo, mediaPath)

		if err != nil {
			return false, err
		}

		if skip {
			return true, nil
		}
	}

	return false, nil
}

func (t scannerTasks) AfterScanAlbum(ctx scanner_task.TaskContext, changedMedia []*models.Media, albumMedia []*models.Media) error {
	return simpleCombinedTasks(ctx, func(ctx scanner_task.TaskContext, task scanner_task.ScannerTask) error {
		return task.AfterScanAlbum(ctx, changedMedia, albumMedia)
	})
}

func (t scannerTasks) AfterMediaFound(ctx scanner_task.TaskContext, media *models.Media, newMedia bool) error {
	return simpleCombinedTasks(ctx, func(ctx scanner_task.TaskContext, task scanner_task.ScannerTask) error {
		return task.AfterMediaFound(ctx, media, newMedia)
	})
}

func (t scannerTasks) BeforeProcessMedia(ctx scanner_task.TaskContext, mediaData *media_encoding.EncodeMediaData) (scanner_task.TaskContext, error) {
	for _, task := range allTasks {
		select {
		case <-ctx.Done():
			return ctx, ctx.Err()
		default:
		}

		var err error
		ctx, err = task.BeforeProcessMedia(ctx, mediaData)
		if err != nil {
			return ctx, err
		}
	}

	return ctx, nil
}

func (t scannerTasks) ProcessMedia(ctx scanner_task.TaskContext, mediaData *media_encoding.EncodeMediaData, mediaCachePath string) ([]*models.MediaURL, error) {
	allNewMedia := make([]*models.MediaURL, 0)

	for _, task := range allTasks {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		newMedia, err := task.ProcessMedia(ctx, mediaData, mediaCachePath)
		if err != nil {
			return []*models.MediaURL{}, err
		}

		allNewMedia = append(allNewMedia, newMedia...)
	}

	return allNewMedia, nil
}

func (t scannerTasks) AfterProcessMedia(ctx scanner_task.TaskContext, mediaData *media_encoding.EncodeMediaData, updatedURLs []*models.MediaURL, mediaIndex int, mediaTotal int) error {
	return simpleCombinedTasks(ctx, func(ctx scanner_task.TaskContext, task scanner_task.ScannerTask) error {
		return task.AfterProcessMedia(ctx, mediaData, updatedURLs, mediaIndex, mediaTotal)
	})
}
