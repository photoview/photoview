package scanner_task

import (
	"io/fs"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/media_encoding"
)

// ScannerTaskBase provides a default "empty" implementation of ScannerTask,
type ScannerTaskBase struct{}

func (t ScannerTaskBase) BeforeScanAlbum(ctx TaskContext) (TaskContext, error) {
	return ctx, nil
}

func (t ScannerTaskBase) AfterScanAlbum(ctx TaskContext, changedMedia []*models.Media, albumMedia []*models.Media) error {
	return nil
}

func (t ScannerTaskBase) MediaFound(ctx TaskContext, fileInfo fs.FileInfo, mediaPath string) (skip bool, err error) {
	return false, nil
}

func (t ScannerTaskBase) AfterMediaFound(ctx TaskContext, media *models.Media, newMedia bool) error {
	return nil
}

func (t ScannerTaskBase) BeforeProcessMedia(ctx TaskContext, mediaData *media_encoding.EncodeMediaData) (TaskContext, error) {
	return ctx, nil
}

func (t ScannerTaskBase) ProcessMedia(ctx TaskContext, mediaData *media_encoding.EncodeMediaData, mediaCachePath string) (updatedURLs []*models.MediaURL, err error) {
	return []*models.MediaURL{}, nil
}

func (t ScannerTaskBase) AfterProcessMedia(ctx TaskContext, mediaData *media_encoding.EncodeMediaData, updatedURLs []*models.MediaURL, mediaIndex int, mediaTotal int) error {
	return nil
}
