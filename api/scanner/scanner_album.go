package scanner

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/face_detection"
	"github.com/photoview/photoview/api/scanner/media_encoding"
	"github.com/photoview/photoview/api/scanner/scanner_task"
	"github.com/photoview/photoview/api/scanner/scanner_tasks"
	"github.com/photoview/photoview/api/scanner/scanner_utils"
	"github.com/photoview/photoview/api/utils"
	"github.com/pkg/errors"
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
	_, err := os.Stat(rootPath)
	if err != nil {
		log.Printf("Warn: invalid root path: '%s'\n%s\n", rootPath, err)
		return false
	}

	return true
}

func ScanAlbum(ctx scanner_task.TaskContext) {

	newCtx, err := scanner_tasks.Tasks.BeforeScanAlbum(ctx)
	if err != nil {
		scanner_utils.ScannerError("before scan album (%s): %s", ctx.GetAlbum().Path, err)
		return
	}
	ctx = newCtx

	// Scan for photos
	albumMedia, err := findMediaForAlbum(ctx)
	if err != nil {
		scanner_utils.ScannerError("find media for album (%s): %s", ctx.GetAlbum().Path, err)
		return
	}

	albumHasChanges := false
	for count, media := range albumMedia {
		didProcess := false

		transactionError := ctx.GetDB().Transaction(func(tx *gorm.DB) error {
			// processing_was_needed, err = ProcessMedia(tx, media)
			didProcess, err = processMedia(ctx, media)
			if err != nil {
				return errors.Wrapf(err, "process media (%s)", media.Path)
			}

			if didProcess {
				albumHasChanges = true
			}

			if err = scanner_tasks.Tasks.AfterProcessMedia(ctx, media, didProcess, count, len(albumMedia)); err != nil {
				return err
			}

			return nil
		})

		if transactionError != nil {
			scanner_utils.ScannerError("begin database transaction: %s", transactionError)
		}

		if didProcess && media.Type == models.MediaTypePhoto {
			go func(media *models.Media) {
				if face_detection.GlobalFaceDetector == nil {
					return
				}
				if err := face_detection.GlobalFaceDetector.DetectFaces(ctx.GetDB(), media); err != nil {
					scanner_utils.ScannerError("Error detecting faces in image (%s): %s", media.Path, err)
				}
			}(media)
		}
	}

	cleanup_errors := CleanupMedia(ctx.GetDB(), ctx.GetAlbum().ID, albumMedia)
	for _, err := range cleanup_errors {
		scanner_utils.ScannerError("delete old media: %s", err)
	}

	if err := scanner_tasks.Tasks.AfterScanAlbum(ctx, albumHasChanges); err != nil {
		scanner_utils.ScannerError("after scan album: %s", err)
	}
}

func findMediaForAlbum(ctx scanner_task.TaskContext) ([]*models.Media, error) {

	albumMedia := make([]*models.Media, 0)

	dirContent, err := ioutil.ReadDir(ctx.GetAlbum().Path)
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

			skip, err := scanner_tasks.Tasks.MediaFound(ctx, item, mediaPath)
			if err != nil {
				return nil, err
			}
			if skip {
				continue
			}

			// Skip the JPEGs that are compressed version of raw files
			counterpartFile := scanForRawCounterpartFile(mediaPath)
			if counterpartFile != nil {
				continue
			}

			err = ctx.GetDB().Transaction(func(tx *gorm.DB) error {
				media, isNewMedia, err := ScanMedia(tx, mediaPath, ctx.GetAlbum().ID, ctx.GetCache())
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
				scanner_utils.ScannerError("Error scanning media for album (%d): %s\n", ctx.GetAlbum().ID, err)
				continue
			}

		}
	}

	return albumMedia, nil
}

func processMedia(ctx scanner_task.TaskContext, media *models.Media) (bool, error) {
	mediaData := media_encoding.EncodeMediaData{
		Media: media,
	}

	_, err := mediaData.ContentType()
	if err != nil {
		return false, errors.Wrapf(err, "get content-type of media (%s)", media.Path)
	}

	// Make sure media cache directory exists
	mediaCachePath, err := makeMediaCacheDir(media)
	if err != nil {
		return false, errors.Wrap(err, "cache directory error")
	}

	return scanner_tasks.Tasks.ProcessMedia(ctx, &mediaData, *mediaCachePath)
}
