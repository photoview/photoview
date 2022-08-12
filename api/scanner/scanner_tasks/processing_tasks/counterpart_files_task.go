package processing_tasks

import (
	"io/fs"
	"path"
	"path/filepath"
	"strings"

	"github.com/photoview/photoview/api/scanner/media_encoding"
	"github.com/photoview/photoview/api/scanner/media_type"
	"github.com/photoview/photoview/api/scanner/scanner_task"
	"github.com/photoview/photoview/api/scanner/scanner_utils"
	"github.com/pkg/errors"
)

type CounterpartFilesTask struct {
	scanner_task.ScannerTaskBase
}

func (t CounterpartFilesTask) MediaFound(ctx scanner_task.TaskContext, fileInfo fs.FileInfo, mediaPath string) (skip bool, err error) {

	// Skip the JPEGs that are compressed version of raw files
	counterpartFile := scanForRawCounterpartFile(mediaPath)
	if counterpartFile != nil {
		return true, nil
	}

	return false, nil
}

func (t CounterpartFilesTask) BeforeProcessMedia(ctx scanner_task.TaskContext, mediaData *media_encoding.EncodeMediaData) (scanner_task.TaskContext, error) {

	mediaType, err := ctx.GetCache().GetMediaType(mediaData.Media.Path)
	if err != nil {
		return ctx, errors.Wrap(err, "scan for counterpart file")
	}

	if !mediaType.IsRaw() {
		return ctx, nil
	}

	counterpartFile := scanForCompressedCounterpartFile(mediaData.Media.Path)
	if counterpartFile != nil {
		mediaData.CounterpartPath = counterpartFile
	}

	return ctx, nil
}

func scanForCompressedCounterpartFile(imagePath string) *string {
	ext := filepath.Ext(imagePath)
	fileExtType, found := media_type.GetExtensionMediaType(ext)

	if found {
		if fileExtType.IsBasicTypeSupported() {
			return nil
		}
	}

	pathWithoutExt := strings.TrimSuffix(imagePath, path.Ext(imagePath))
	for _, ext := range media_type.TypeJpeg.FileExtensions() {
		testPath := pathWithoutExt + ext
		if scanner_utils.FileExists(testPath) {
			return &testPath
		}
	}

	return nil
}

func scanForRawCounterpartFile(imagePath string) *string {
	ext := filepath.Ext(imagePath)
	fileExtType, found := media_type.GetExtensionMediaType(ext)

	if found {
		if !fileExtType.IsBasicTypeSupported() {
			return nil
		}
	}

	rawPath := media_type.RawCounterpart(imagePath)
	if rawPath != nil {
		return rawPath
	}

	return nil
}
