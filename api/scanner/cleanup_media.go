package scanner

import (
	"os"
	"path"
	"strconv"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

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
		query.Where("NOT id IN ?", albumMediaIds)
	}

	if err := query.Find(&mediaList).Error; err != nil {
		return []error{errors.Wrap(err, "get media files to be deleted from database")}
	}

	deleteErrors := make([]error, 0)

	mediaIDs := make([]int, 0)
	for _, media := range mediaList {

		mediaIDs = append(mediaIDs, media.ID)
		cachePath := path.Join(PhotoCache(), strconv.Itoa(int(albumId)), strconv.Itoa(int(media.ID)))
		err := os.RemoveAll(cachePath)
		if err != nil {
			deleteErrors = append(deleteErrors, errors.Wrapf(err, "delete unused cache folder (%s)", cachePath))
		}

	}

	if len(mediaIDs) > 0 {
		if err := db.Where("id IN ?", mediaIDs).Delete(models.Media{}).Error; err != nil {
			deleteErrors = append(deleteErrors, errors.Wrap(err, "delete old media from database"))
		}
	}

	return deleteErrors
}

func deleteOldUserAlbums(db *gorm.DB, scannedAlbums []*models.Album, user *models.User) []error {
	if len(scannedAlbums) == 0 {
		return nil
	}

	scannedAlbumIDs := make([]interface{}, len(scannedAlbums))
	for i, album := range scannedAlbums {
		scannedAlbumIDs[i] = album.ID
	}

	// Delete old albums
	var albums []models.Album

	userAlbumIDs := make([]int, len(user.Albums))
	for i, album := range user.Albums {
		userAlbumIDs[i] = album.ID
	}

	query := db.
		Where("id IN (?)", userAlbumIDs).
		Where("id NOT IN (?)", scannedAlbumIDs)

	if err := query.Find(&albums).Error; err != nil {
		return []error{errors.Wrap(err, "get albums to be deleted from database")}
	}

	deleteErrors := make([]error, 0)

	deleteAlbumIDs := make([]int, len(albums))
	for i, album := range albums {
		deleteAlbumIDs[i] = album.ID
		cachePath := path.Join(PhotoCache(), strconv.Itoa(int(album.ID)))
		err := os.RemoveAll(cachePath)
		if err != nil {
			deleteErrors = append(deleteErrors, errors.Wrapf(err, "delete unused cache folder (%s)", cachePath))
		}
	}

	if err := db.Where("id IN ?", deleteAlbumIDs).Delete(models.Album{}).Error; err != nil {
		ScannerError("Could not delete old albums from database:\n%s\n", err)
		deleteErrors = append(deleteErrors, errors.Wrap(err, "delete old albums from database"))
	}

	return deleteErrors
}
