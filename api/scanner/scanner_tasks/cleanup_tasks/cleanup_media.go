package cleanup_tasks

import (
	"os"
	"path"
	"strconv"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/face_detection"
	"github.com/photoview/photoview/api/scanner/scanner_utils"
	"github.com/photoview/photoview/api/utils"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// CleanupMedia removes media entries from the database that are no longer present on the filesystem
func CleanupMedia(db *gorm.DB, albumId int, albumMedia []*models.Media) []error {
	albumMediaIds := make([]int, len(albumMedia))
	for i, media := range albumMedia {
		albumMediaIds[i] = media.ID
	}

	// Will get from database
	var mediaList []models.Media

	query := db.Where("album_id = ?", albumId)

	// Select media from database that was not found on hard disk
	if len(albumMedia) > 0 {
		query = query.Where("NOT id IN (?)", albumMediaIds)
	}

	if err := query.Find(&mediaList).Error; err != nil {
		return []error{errors.Wrap(err, "get media files to be deleted from database")}
	}

	deleteErrors := make([]error, 0)

	mediaIDs := make([]int, 0)
	for _, media := range mediaList {

		mediaIDs = append(mediaIDs, media.ID)
		cachePath := path.Join(utils.MediaCachePath(), strconv.Itoa(int(albumId)), strconv.Itoa(int(media.ID)))
		err := os.RemoveAll(cachePath)
		if err != nil {
			deleteErrors = append(deleteErrors, errors.Wrapf(err, "delete unused cache folder (%s)", cachePath))
		}

	}

	if len(mediaIDs) > 0 {
		if err := db.Where("id IN (?)", mediaIDs).Delete(models.Media{}).Error; err != nil {
			deleteErrors = append(deleteErrors, errors.Wrap(err, "delete old media from database"))
		}

		// Reload faces after deleting media
		if face_detection.GlobalFaceDetector != nil {
			if err := face_detection.GlobalFaceDetector.ReloadFacesFromDatabase(db); err != nil {
				deleteErrors = append(deleteErrors, errors.Wrap(err, "reload faces from database"))
			}
		}
	}

	return deleteErrors
}

// DeleteOldUserAlbums finds and deletes old albums in the database and cache that does not exist on the filesystem anymore.
func DeleteOldUserAlbums(db *gorm.DB, scannedAlbums []*models.Album, user *models.User) []error {
	if len(scannedAlbums) == 0 {
		return nil
	}

	scannedAlbumIDs := make([]interface{}, len(scannedAlbums))
	for i, album := range scannedAlbums {
		scannedAlbumIDs[i] = album.ID
	}

	// Old albums to be deleted
	var deleteAlbums []models.Album

	// Find old albums in database
	query := db.
		Select("albums.*").
		Table("user_albums").
		Joins("JOIN albums ON user_albums.album_id = albums.id").
		Where("user_id = ?", user.ID).
		Where("album_id NOT IN (?)", scannedAlbumIDs)

	if err := query.Find(&deleteAlbums).Error; err != nil {
		return []error{errors.Wrap(err, "get albums to be deleted from database")}
	}

	if len(deleteAlbums) == 0 {
		return []error{}
	}

	deleteErrors := make([]error, 0)

	// Delete old albums from cache
	deleteAlbumIDs := make([]int, len(deleteAlbums))
	for i, album := range deleteAlbums {
		deleteAlbumIDs[i] = album.ID
		cachePath := path.Join(utils.MediaCachePath(), strconv.Itoa(int(album.ID)))
		err := os.RemoveAll(cachePath)
		if err != nil {
			deleteErrors = append(deleteErrors, errors.Wrapf(err, "delete unused cache folder (%s)", cachePath))
		}
	}

	// Delete old albums from database
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("album_id IN (?)", deleteAlbumIDs).Delete(&models.UserAlbums{}).Error; err != nil {
			return err
		}

		// A grant can also reference one of these albums as its source
		// without living on it (e.g. a descendant a few levels down that
		// isn't itself stale) - find those before deleting, so their
		// UserAlbums row can be recomputed afterwards rather than left
		// stale with a level that included a source about to disappear.
		var strandedGrants []models.UserAlbumGrant
		if err := tx.
			Where("source_album_id IN (?) AND album_id NOT IN (?)", deleteAlbumIDs, deleteAlbumIDs).
			Find(&strandedGrants).Error; err != nil {
			return err
		}

		// Also delete the provenance rows backing those materialized grants -
		// left behind, they'd keep referencing an album that's about to be
		// deleted, either as the grant's own album or as its source.
		if err := tx.
			Where("album_id IN (?) OR source_album_id IN (?)", deleteAlbumIDs, deleteAlbumIDs).
			Delete(&models.UserAlbumGrant{}).Error; err != nil {
			return err
		}

		if err := tx.Where("id IN (?)", deleteAlbumIDs).Delete(models.Album{}).Error; err != nil {
			return err
		}

		for _, g := range strandedGrants {
			if err := models.RecomputeUserAlbums(tx, g.UserID, []int{g.AlbumID}); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		scanner_utils.ScannerError(nil, "Could not delete old albums from database:\n%s\n", err)
		deleteErrors = append(deleteErrors, err)
	}

	// Reload faces after deleting albums
	if face_detection.GlobalFaceDetector != nil {
		if err := face_detection.GlobalFaceDetector.ReloadFacesFromDatabase(db); err != nil {
			deleteErrors = append(deleteErrors, err)
		}
	}

	return deleteErrors
}
