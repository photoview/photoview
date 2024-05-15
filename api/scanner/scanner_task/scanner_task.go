package scanner_task

import (
	"context"
	"database/sql"
	"flag"
	"io/fs"

	"github.com/kkovaletp/photoview/api/graphql/models"
	"github.com/kkovaletp/photoview/api/scanner/media_encoding"
	"github.com/kkovaletp/photoview/api/scanner/scanner_cache"
	"gorm.io/gorm"
)

// ScannerTask is an interface for a task to be performed as a part of the scanner pipeline
type ScannerTask interface {
	// BeforeScanAlbum will run at the beginning of the scan task.
	// New values can be stored in the returned TaskContext that will live throughout the lifetime of the task.
	BeforeScanAlbum(ctx TaskContext) (TaskContext, error)

	// AfterScanAlbum will run at the end of the scan task.
	AfterScanAlbum(ctx TaskContext, changedMedia []*models.Media, albumMedia []*models.Media) error

	// MediaFound will run for each media file found on the filesystem.
	// It will run even when the media is already present in the database.
	// If the returned skip value is true, the media will be skipped and further steps will not be executed for the given file.
	MediaFound(ctx TaskContext, fileInfo fs.FileInfo, mediaPath string) (skip bool, err error)

	// AfterMediaFound will run each media file after is has been saved to the database, but not processed yet.
	// It will run even when the media is already present in the database, in that case `newMedia` will be true.
	AfterMediaFound(ctx TaskContext, media *models.Media, newMedia bool) error

	BeforeProcessMedia(ctx TaskContext, mediaData *media_encoding.EncodeMediaData) (TaskContext, error)
	ProcessMedia(ctx TaskContext, mediaData *media_encoding.EncodeMediaData, mediaCachePath string) (updatedURLs []*models.MediaURL, err error)
	AfterProcessMedia(ctx TaskContext, mediaData *media_encoding.EncodeMediaData, updatedURLs []*models.MediaURL, mediaIndex int, mediaTotal int) error
}

type TaskContext struct {
	ctx context.Context
}

func NewTaskContext(parent context.Context, db *gorm.DB, album *models.Album, cache *scanner_cache.AlbumScannerCache) TaskContext {
	ctx := TaskContext{ctx: parent}
	ctx = ctx.WithValue(taskCtxKeyAlbum, album)
	ctx = ctx.WithValue(taskCtxKeyAlbumCache, cache)
	ctx = ctx.WithDB(db)

	return ctx
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

func (c TaskContext) DatabaseTransaction(transFunc func(ctx TaskContext) error, opts ...*sql.TxOptions) error {
	return c.GetDB().Transaction(func(tx *gorm.DB) error {
		return transFunc(c.WithDB(tx))
	}, opts...)
}

func (c TaskContext) WithValue(key, val interface{}) TaskContext {
	return TaskContext{
		ctx: context.WithValue(c.ctx, key, val),
	}
}

func (c TaskContext) Value(key interface{}) interface{} {
	return c.ctx.Value(key)
}

func (c TaskContext) WithDB(db *gorm.DB) TaskContext {
	// Allow db to be nil in tests
	if db == nil && flag.Lookup("test.v") != nil {
		return c
	}

	return c.WithValue(taskCtxKeyDatabase, db.WithContext(c.ctx))
}

func (c TaskContext) Done() <-chan struct{} {
	return c.ctx.Done()
}

func (c TaskContext) Err() error {
	return c.ctx.Err()
}
