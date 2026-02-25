package scanner

import (
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/media_encoding"
	"github.com/photoview/photoview/api/scanner/scanner_task"
	"github.com/photoview/photoview/api/scanner/scanner_tasks"
	"github.com/pkg/errors"
)

func scanMedia(ctx scanner_task.TaskContext, media *models.Media, mediaData *media_encoding.EncodeMediaData, mediaIndex int, mediaTotal int) error {
	cacheFs := ctx.GetCacheFS()

	newCtx, err := scanner_tasks.Tasks.BeforeProcessMedia(ctx, mediaData)
	if err != nil {
		return errors.Wrapf(err, "before process media (%s)", media.Path)
	}

	mediaCachePath, err := media.CachePath(cacheFs)
	if err != nil {
		return errors.Wrapf(err, "cache directory error (%s)", media.Path)
	}

	transactionError := newCtx.DatabaseTransaction(func(ctx scanner_task.TaskContext) error {
		updatedURLs, err := scanner_tasks.Tasks.ProcessMedia(newCtx, mediaData, mediaCachePath)
		if err != nil {
			return errors.Wrapf(err, "process media (%s)", media.Path)
		}

		if err = scanner_tasks.Tasks.AfterProcessMedia(newCtx, mediaData, updatedURLs, mediaIndex, mediaTotal); err != nil {
			return errors.Wrap(err, "after process media")
		}

		return nil
	})

	if transactionError != nil {
		return errors.Wrap(transactionError, "process media database transaction")
	}

	return nil
}
