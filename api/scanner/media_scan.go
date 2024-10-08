package scanner

import (
	"fmt"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/media_encoding"
	"github.com/photoview/photoview/api/scanner/scanner_task"
	"github.com/photoview/photoview/api/scanner/scanner_tasks"
)

func scanMedia(ctx scanner_task.TaskContext, media *models.Media, mediaData *media_encoding.EncodeMediaData, mediaIndex int, mediaTotal int) error {
	newCtx, err := scanner_tasks.Tasks.BeforeProcessMedia(ctx, mediaData)
	if err != nil {
		return fmt.Errorf("before process media (%s): %w", media.Path, err)
	}

	mediaCachePath, err := media.CachePath()
	if err != nil {
		return fmt.Errorf("cache directory error (%s): %w", media.Path, err)
	}

	transactionError := newCtx.DatabaseTransaction(func(ctx scanner_task.TaskContext) error {
		updatedURLs, err := scanner_tasks.Tasks.ProcessMedia(newCtx, mediaData, mediaCachePath)
		if err != nil {
			return fmt.Errorf("process media (%s): %w", media.Path, err)
		}

		if err = scanner_tasks.Tasks.AfterProcessMedia(newCtx, mediaData, updatedURLs, mediaIndex, mediaTotal); err != nil {
			return fmt.Errorf("after process media: %w", err)
		}

		return nil
	})

	if transactionError != nil {
		return fmt.Errorf("process media database transaction: %w", transactionError)
	}

	return nil
}
