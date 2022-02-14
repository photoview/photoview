package scanner_task

import (
	"context"
	"io/fs"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/media_encoding"
	"github.com/photoview/photoview/api/scanner/scanner_cache"
	"gorm.io/gorm"
)

// ScannerTask is an interface for a task to be performed as a part of the scanner pipeline
type ScannerTask interface {
	// BeforeScanAlbum will run at the beginning of the scan task.
	// New values can be stored in the returned TaskContext that will live throughout the lifetime of the task.
	BeforeScanAlbum(ctx TaskContext) (TaskContext, error)
	AfterScanAlbum(ctx TaskContext, albumHadChanges bool) error

	MediaFound(ctx TaskContext, fileInfo fs.FileInfo, mediaPath string) (skip bool, err error)
	AfterMediaFound(ctx TaskContext, media *models.Media, newMedia bool) error

	BeforeProcessMedia(ctx TaskContext, media *models.Media) (TaskContext, error)
	ProcessMedia(ctx TaskContext, mediaData *media_encoding.EncodeMediaData, mediaCachePath string) (didProcess bool, err error)
	AfterProcessMedia(ctx TaskContext, media *models.Media, didProcess bool, mediaIndex int, mediaTotal int) error
}

type TaskContext struct {
	ctx context.Context
}

func NewTaskContext(parent context.Context, db *gorm.DB, album *models.Album, cache *scanner_cache.AlbumScannerCache) TaskContext {
	ctx := parent
	ctx = context.WithValue(ctx, taskCtxKeyAlbum, album)
	ctx = context.WithValue(ctx, taskCtxKeyAlbumCache, cache)
	ctx = context.WithValue(ctx, taskCtxKeyDatabase, db.WithContext(ctx))

	return TaskContext{
		ctx: ctx,
	}
}

type taskCtxKeyType string

const (
	taskCtxKeyAlbum      taskCtxKeyType = "task_album"
	taskCtxKeyAlbumCache taskCtxKeyType = "task_album_cache"
	taskCtxKeyDatabase   taskCtxKeyType = "task_database"
)

func (c TaskContext) GetAlbum() *models.Album {
	return c.ctx.Value(taskCtxKeyAlbum).(*models.Album)
}

func (c TaskContext) GetCache() *scanner_cache.AlbumScannerCache {
	return c.ctx.Value(taskCtxKeyAlbumCache).(*scanner_cache.AlbumScannerCache)
}

func (c TaskContext) GetDB() *gorm.DB {
	return c.ctx.Value(taskCtxKeyDatabase).(*gorm.DB)
}

func (c TaskContext) WithValue(key, val interface{}) TaskContext {
	return TaskContext{
		ctx: context.WithValue(c.ctx, key, val),
	}
}

func (c TaskContext) Value(key interface{}) interface{} {
	return c.ctx.Value(key)
}
