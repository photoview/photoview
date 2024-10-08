package scanner

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/media_encoding"
	"github.com/photoview/photoview/api/scanner/scanner_task"
	"github.com/photoview/photoview/api/scanner/scanner_tasks"
	"github.com/photoview/photoview/api/scanner/scanner_utils"
	"github.com/photoview/photoview/api/utils"
	"gorm.io/gorm"
)

func NewRootAlbum(db *gorm.DB, rootPath string, owner *models.User) (*models.Album, error) {

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
			return nil, fmt.Errorf("user already owns a path containing this path: %s", rootPath)
		}

		if err := db.Model(&owner).Association("Albums").Append(&album); err != nil {
			return nil, fmt.Errorf("add owner to already existing album: %w", err)
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

var ErrorInvalidRootPath = fmt.Errorf("invalid root path")

func ValidRootPath(rootPath string) bool {
	_, err := os.Stat(rootPath)
	if err != nil {
		log.Printf("Warn: invalid root path: '%s'\n%s\n", rootPath, err)
		return false
	}

	return true
}

func ScanAlbum(ctx scanner_task.TaskContext) error {
	newCtx, err := scanner_tasks.Tasks.BeforeScanAlbum(ctx)
	if err != nil {
		return fmt.Errorf("before scan album (%s): %w", ctx.GetAlbum().Path, err)
	}
	ctx = newCtx

	// Scan for photos
	albumMedia, err := findMediaForAlbum(ctx)
	if err != nil {
		return fmt.Errorf("find media for album (%s): %w", ctx.GetAlbum().Path, err)
	}

	changedMedia := make([]*models.Media, 0)
	for i, media := range albumMedia {
		mediaData := media_encoding.NewEncodeMediaData(media)

		if err := scanMedia(ctx, media, &mediaData, i, len(albumMedia)); err != nil {
			scanner_utils.ScannerError("Error scanning media for album (%d) file (%s): %s\n",
				ctx.GetAlbum().ID, media.Path, err)
		}
	}

	if err := scanner_tasks.Tasks.AfterScanAlbum(ctx, changedMedia, albumMedia); err != nil {
		return fmt.Errorf("after scan album: %w", err)
	}

	return nil
}

func findMediaForAlbum(ctx scanner_task.TaskContext) ([]*models.Media, error) {

	albumMedia := make([]*models.Media, 0)

	dirContent, err := os.ReadDir(ctx.GetAlbum().Path)
	if err != nil {
		return nil, err
	}

	for _, item := range dirContent {
		mediaPath := path.Join(ctx.GetAlbum().Path, item.Name())

		isDirSymlink, err := utils.IsDirSymlink(mediaPath)
		if err != nil {
			log.Printf("Cannot detect whether %s is symlink to a directory. Pretending it is not", mediaPath)
			isDirSymlink = false
		}

		if !item.IsDir() && !isDirSymlink && ctx.GetCache().IsPathMedia(mediaPath) {
			itemInfo, err := item.Info()
			if err != nil {
				return nil, err
			}
			skip, err := scanner_tasks.Tasks.MediaFound(ctx, itemInfo, mediaPath)
			if err != nil {
				return nil, err
			}
			if skip {
				continue
			}

			err = ctx.DatabaseTransaction(func(ctx scanner_task.TaskContext) error {
				media, isNewMedia, err := ScanMedia(ctx.GetDB(), mediaPath, ctx.GetAlbum().ID, ctx.GetCache())
				if err != nil {
					return fmt.Errorf("scanning media error (%s): %w", mediaPath, err)
				}

				if err = scanner_tasks.Tasks.AfterMediaFound(ctx, media, isNewMedia); err != nil {
					return err
				}

				albumMedia = append(albumMedia, media)

				return nil
			})

			if err != nil {
				scanner_utils.ScannerError("Error scanning media for album (%d): %s\n", ctx.GetAlbum().ID, err)
				continue
			}
		}

	}

	return albumMedia, nil
}

func processMedia(ctx scanner_task.TaskContext, mediaData *media_encoding.EncodeMediaData) ([]*models.MediaURL, error) {

	// Make sure media cache directory exists
	mediaCachePath, err := mediaData.Media.CachePath()
	if err != nil {
		return []*models.MediaURL{}, fmt.Errorf("cache directory error: %w", err)
	}

	return scanner_tasks.Tasks.ProcessMedia(ctx, mediaData, mediaCachePath)
}
