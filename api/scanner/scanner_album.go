package scanner

import (
	"fmt"
	"os"
	"path"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/log"
	"github.com/photoview/photoview/api/scanner/media_encoding"
	"github.com/photoview/photoview/api/scanner/scanner_task"
	"github.com/photoview/photoview/api/scanner/scanner_tasks"
	"github.com/photoview/photoview/api/scanner/scanner_utils"
	"github.com/photoview/photoview/api/utils"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"gorm.io/gorm"
)

func NewRootAlbum(db *gorm.DB, fs afero.Fs, rootPath string, owner *models.User) (*models.Album, error) {

	if !ValidRootPath(fs, rootPath) {
		return nil, ErrorInvalidRootPath
	}

	if !path.IsAbs(rootPath) {
		if _, ok := fs.(*afero.OsFs); ok {
			wd, err := os.Getwd()
			if err != nil {
				return nil, err
			}
			rootPath = path.Join(wd, rootPath)
		} else {
			rootPath = path.Clean(rootPath)
		}
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

func ValidRootPath(fs afero.Fs, rootPath string) bool {
	_, err := fs.Stat(rootPath)
	if err != nil {
		log.Warn(nil, "invalid root path", "root_path", rootPath, "error", err)
		return false
	}

	return true
}

func ScanAlbum(ctx scanner_task.TaskContext) error {
	fs := ctx.GetFileFS()

	newCtx, err := scanner_tasks.Tasks.BeforeScanAlbum(ctx)
	if err != nil {
		return errors.Wrapf(err, "before scan album (%s)", ctx.GetAlbum().Path)
	}
	ctx = newCtx

	// Scan for photos
	albumMedia, err := findMediaForAlbum(ctx)
	if err != nil {
		return errors.Wrapf(err, "find media for album (%s): %s", ctx.GetAlbum().Path, err)
	}

	changedMedia := make([]*models.Media, 0)
	for i, media := range albumMedia {
		// Download to temporary local path if needed
		media.LocalPath, err = scanner_utils.DownloadToLocalIfNeeded(fs, media.Path)
		if err != nil {
			return errors.Wrapf(err, "could not download local media path: %s", media.Path)
		}

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

func findMediaForAlbum(ctx scanner_task.TaskContext) ([]*models.Media, error) {
	fs := ctx.GetFileFS()

	albumMedia := make([]*models.Media, 0)

	dirContent, err := afero.ReadDir(fs, ctx.GetAlbum().Path)
	if err != nil {
		return nil, err
	}

	for _, item := range dirContent {
		mediaPath := path.Join(ctx.GetAlbum().Path, item.Name())
		log.Info(ctx, "Check the media", "media_path", mediaPath)

		isDirSymlink, err := utils.IsDirSymlink(fs, mediaPath)
		if err != nil {
			log.Warn(ctx, "Cannot detect whether the path is symlink to a directory. Pretending it is not", "media_path", mediaPath)
			isDirSymlink = false
		}

		// FIXME: should we download to local path here?

		if !item.IsDir() && !isDirSymlink && ctx.GetCache().IsPathMedia(mediaPath) {
			skip, err := scanner_tasks.Tasks.MediaFound(ctx, item, mediaPath, mediaPath)
			if err != nil {
				return nil, err
			}
			if skip {
				continue
			}

			err = ctx.DatabaseTransaction(func(ctx scanner_task.TaskContext) error {
				media, isNewMedia, err := ScanMedia(ctx.GetDB(), fs, mediaPath, mediaPath, ctx.GetAlbum().ID, ctx.GetCache())
				if err != nil {
					return errors.Wrapf(err, "scanning media error (%s)", mediaPath)
				}

				if err = scanner_tasks.Tasks.AfterMediaFound(ctx, media, isNewMedia); err != nil {
					return err
				}

				albumMedia = append(albumMedia, media)

				return nil
			})

			if err != nil {
				scanner_utils.ScannerError(ctx, "Error scanning media for album (%d): %s\n", ctx.GetAlbum().ID, err)
				continue
			}
		}

	}

	return albumMedia, nil
}
