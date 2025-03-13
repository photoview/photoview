package processing_tasks

import (
	"fmt"
	"io/fs"

	"github.com/photoview/photoview/api/scanner/media_encoding"
	"github.com/photoview/photoview/api/scanner/media_type"
	"github.com/photoview/photoview/api/scanner/scanner_task"
	"github.com/photoview/photoview/api/utils"
)

type CounterpartFilesTask struct {
	scanner_task.ScannerTaskBase
}

func (t CounterpartFilesTask) MediaFound(ctx scanner_task.TaskContext, fileInfo fs.FileInfo, mediaPath string) (skip bool, err error) {
	fileType := media_type.GetMediaType(mediaPath)

	if !fileType.IsSupported() {
		return true, nil
	}

	if utils.EnvDisableRawProcessing.GetBool() {
		if !fileType.IsWebCompatible() {
			return true, nil
		}

		// Don't skip the JPEGs if raw processing is disabled. Treat them as standalone files.
		return false, nil
	}

	if fileType.IsWebCompatible() {
		_, existed := media_type.FindRawCounterpart(mediaPath)
		if existed {
			return true, nil
		}
	}

	return false, nil
}

func (t CounterpartFilesTask) BeforeProcessMedia(ctx scanner_task.TaskContext, mediaData *media_encoding.EncodeMediaData) (scanner_task.TaskContext, error) {

	mediaType, err := ctx.GetCache().GetMediaType(mediaData.Media.Path)
	if err != nil {
		return ctx, fmt.Errorf("scan for counterpart file error: %w", err)
	}

	if mediaType.IsWebCompatible() {
		return ctx, nil
	}

	counterpartFile, ok := media_type.FindWebCounterpart(mediaData.Media.Path)
	if !ok {
		return ctx, nil
	}

	mediaData.CounterpartPath = &counterpartFile

	return ctx, nil
}
