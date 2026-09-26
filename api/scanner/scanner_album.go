package scanner

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/log"
	"github.com/photoview/photoview/api/scanner/media_encoding"
	"github.com/photoview/photoview/api/scanner/media_type"
	"github.com/photoview/photoview/api/scanner/scanner_task"
	"github.com/photoview/photoview/api/scanner/scanner_tasks"
	"github.com/photoview/photoview/api/scanner/scanner_utils"
	"github.com/photoview/photoview/api/utils"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func NewRootAlbum(db *gorm.DB, rootPath string, owner *models.User) (*models.Album, error) {
	rootPath = filepath.Clean(rootPath)

	if !ValidRootPath(rootPath) {
		return nil, ErrorInvalidRootPath
	}

	if !path.IsAbs(rootPath) {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}

		rootPath = path.Join(wd, rootPath)
	}

	owners := []models.User{
		*owner,
	}

	var matchedAlbums []models.Album
	if err := db.Where("path_hash = ?", models.MD5Hash(rootPath)).Find(&matchedAlbums).Error; err != nil {
		return nil, err
	}

	if len(matchedAlbums) > 0 {
		album := matchedAlbums[0]

		var matchedUserAlbumCount int64
		if err := db.Table("user_albums").Where("user_id = ?", owner.ID).Where("album_id = ?", album.ID).Count(&matchedUserAlbumCount).Error; err != nil {
			return nil, err
		}

		if matchedUserAlbumCount > 0 {
			return nil, errors.New(fmt.Sprintf("user already owns a path containing this path: %s", rootPath))
		}

		if err := db.Model(&owner).Association("Albums").Append(&album); err != nil {
			return nil, errors.Wrap(err, "add owner to already existing album")
		}

		return &album, nil
	} else {
		album := models.Album{
			Title:  path.Base(rootPath),
			Path:   rootPath,
			Owners: owners,
		}

		if err := db.Create(&album).Error; err != nil {
			return nil, err
		}

		return &album, nil
	}
}

var ErrorInvalidRootPath = errors.New("invalid root path")

func ValidRootPath(rootPath string) bool {
	resolvedPath, err := filepath.EvalSymlinks(rootPath)
	if err != nil {
		log.Warn(nil, "invalid root path", "root_path", rootPath, "error", err)
		return false
	}

	// Confirm the resolved path is actually a directory, not a file.
	info, err := os.Stat(resolvedPath)
	if err != nil {
		log.Warn(nil, "invalid root path after symlink resolution",
			"root_path", rootPath, "resolved_path", resolvedPath, "error", err)
		return false
	}
	if !info.IsDir() {
		log.Warn(nil, "root path is not a directory", "root_path", rootPath)
		return false
	}

	return true
}

func ScanAlbum(ctx scanner_task.TaskContext) error {
	newCtx, err := scanner_tasks.Tasks.BeforeScanAlbum(ctx)
	if err != nil {
		return errors.Wrapf(err, "before scan album (%s)", ctx.GetAlbum().Path)
	}
	ctx = newCtx

	// Scan for photos
	albumMedia, unscannedPaths, err := findMediaForAlbum(ctx)
	if err != nil {
		return errors.Wrapf(err, "find media for album (%s): %s", ctx.GetAlbum().Path, err)
	}

	// The after-scan cleanup deletes every media row missing from albumMedia.
	// A file that failed to scan is missing for a reason other than being gone
	// from disk, so tell the cleanup to keep it.
	ctx = ctx.WithUnscannedMediaPaths(unscannedPaths)

	changedMedia := make([]*models.Media, 0)
	for i, media := range albumMedia {
		mediaData := media_encoding.NewEncodeMediaData(media)

		if err := scanMedia(ctx, media, &mediaData, i, len(albumMedia)); err != nil {
			scanner_utils.ScannerError(ctx, "Error scanning media for album (%d) file (%s): %s\n", ctx.GetAlbum().ID, media.Path, err)
		}
	}

	if err := scanner_tasks.Tasks.AfterScanAlbum(ctx, changedMedia, albumMedia); err != nil {
		return errors.Wrap(err, "after scan album")
	}

	return nil
}

// mediaScanAttempts is how often one file's database transaction is tried
// before the file counts as unscanned. Most failures a healthy database
// produces here are transient - a lock timeout, a dropped connection, a
// deadlock between the scanner's workers - and a second attempt costs
// nothing when it succeeds. Only a failure that survives all attempts falls
// through to the caller, which then keeps the media row rather than treating
// the file as gone.
const mediaScanAttempts = 3

// mediaScanRetryDelay is the wait before the next attempt, multiplied by the
// attempt number, so the database gets a moment to recover rather than being
// hit again immediately.
const mediaScanRetryDelay = 100 * time.Millisecond

// worthRetrying reports whether a second attempt could end differently. A
// file whose media type cannot be determined, or one that has gone or cannot
// be read, fails the same way every time: waiting for it only slows the scan
// down. Everything else - a lock timeout, a dropped connection, a deadlock
// between the scanner's workers - is worth another attempt.
func worthRetrying(err error) bool {
	return !errors.Is(err, media_type.ErrUnknownType) &&
		!errors.Is(err, fs.ErrNotExist) &&
		!errors.Is(err, fs.ErrPermission)
}

// scanMediaFile runs one file's scan in a database transaction, retrying a
// failed transaction a few times. The media is returned only after the
// transaction has actually committed: a commit that fails after the callback
// returned would otherwise count the media twice on the next attempt.
func scanMediaFile(ctx scanner_task.TaskContext, mediaPath string) (*models.Media, error) {
	var lastErr error

	for attempt := 1; attempt <= mediaScanAttempts; attempt++ {
		var scanned *models.Media

		lastErr = ctx.DatabaseTransaction(func(ctx scanner_task.TaskContext) error {
			media, isNewMedia, err := ScanMedia(ctx.GetDB(), mediaPath, ctx.GetAlbum().ID, ctx.GetCache())
			if err != nil {
				return errors.Wrapf(err, "scanning media error (%s)", mediaPath)
			}

			if err = scanner_tasks.Tasks.AfterMediaFound(ctx, media, isNewMedia); err != nil {
				return err
			}

			scanned = media

			return nil
		})

		if lastErr == nil {
			return scanned, nil
		}

		if attempt == mediaScanAttempts || !worthRetrying(lastErr) {
			break
		}

		log.Warn(ctx, "Media scan transaction failed, retrying",
			"media_path", mediaPath, "attempt", attempt, "error", lastErr)

		// A cancelled scan must not sit out its retries.
		select {
		case <-ctx.Done():
			return nil, lastErr
		case <-time.After(time.Duration(attempt) * mediaScanRetryDelay):
		}
	}

	return nil, lastErr
}

// findMediaForAlbum returns the media found in the album's directory, and the
// paths of media files that are on disk but could not be scanned.
func findMediaForAlbum(ctx scanner_task.TaskContext) ([]*models.Media, []string, error) {

	albumMedia := make([]*models.Media, 0)
	unscannedPaths := make([]string, 0)

	dirContent, err := os.ReadDir(ctx.GetAlbum().Path)
	if err != nil {
		return nil, nil, err
	}

	for _, item := range dirContent {
		mediaPath := path.Join(ctx.GetAlbum().Path, item.Name())
		log.Info(ctx, "Check the media", "media_path", mediaPath)

		isDirSymlink, err := utils.IsDirSymlink(mediaPath)
		if err != nil {
			log.Warn(ctx, "Cannot detect whether the path is symlink to a directory. Pretending it is not", "media_path", mediaPath)
			isDirSymlink = false
		}

		if !item.IsDir() && !isDirSymlink && ctx.GetCache().IsPathMedia(mediaPath) {
			itemInfo, err := item.Info()
			if err != nil {
				return nil, nil, err
			}
			skip, err := scanner_tasks.Tasks.MediaFound(ctx, itemInfo, mediaPath)
			if err != nil {
				return nil, nil, err
			}
			if skip {
				continue
			}

			media, err := scanMediaFile(ctx, mediaPath)
			if err != nil {
				scanner_utils.ScannerError(ctx, "Error scanning media for album (%d): %s\n", ctx.GetAlbum().ID, err)
				unscannedPaths = append(unscannedPaths, mediaPath)
				continue
			}

			albumMedia = append(albumMedia, media)
		}

	}

	return albumMedia, unscannedPaths, nil
}

func processMedia(ctx scanner_task.TaskContext, mediaData *media_encoding.EncodeMediaData) ([]*models.MediaURL, error) {

	// Make sure media cache directory exists
	mediaCachePath, err := mediaData.Media.CachePath()
	if err != nil {
		return []*models.MediaURL{}, errors.Wrap(err, "cache directory error")
	}

	return scanner_tasks.Tasks.ProcessMedia(ctx, mediaData, mediaCachePath)
}
