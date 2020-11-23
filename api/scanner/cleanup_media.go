package scanner

import (
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/viktorstrate/photoview/api/graphql/models"
	"gorm.io/gorm"
)

func CleanupMedia(db *gorm.DB, albumId uint, albumMedia []*models.Media) []error {
	albumMediaIds := make([]uint, len(albumMedia))
	for i, media := range albumMedia {
		albumMediaIds[i] = media.ID
	}

	// Will get from database
	var mediaList []models.Media

	db.Where("album_id = ?", albumId)

	// Select media from database that was not found on hard disk
	if len(albumMedia) > 0 {
		db.Not(albumMediaIds)
	}

	if err := db.Find(&mediaList).Error; err != nil {
		return []error{errors.Wrap(err, "get media files to be deleted from database")}
	}

	deleteErrors := make([]error, 0)

	for _, media := range mediaList {

		// deletedMediaIDs = append(deletedMediaIDs, media.ID)
		cachePath := path.Join(PhotoCache(), strconv.Itoa(int(albumId)), strconv.Itoa(int(media.ID)))
		err := os.RemoveAll(cachePath)
		if err != nil {
			deleteErrors = append(deleteErrors, errors.Wrapf(err, "delete unused cache folder (%s)", cachePath))
		}

	}

	if err := db.Delete(&mediaList).Error; err != nil {
		deleteErrors = append(deleteErrors, errors.Wrap(err, "delete old media from database"))
	}

	return deleteErrors
}

func deleteOldUserAlbums(db *gorm.DB, scannedAlbums []*models.Album, user *models.User) []error {
	if len(scannedAlbums) == 0 {
		return nil
	}

	albumPaths := make([]interface{}, len(scannedAlbums))
	for i, album := range scannedAlbums {
		albumPaths[i] = album.Path
	}

	// Delete old albums
	album_args := make([]interface{}, 0)
	album_args = append(album_args, user.ID)
	album_args = append(album_args, albumPaths...)

	var albums []models.Album

	albums_questions := strings.Repeat("MD5(?),", len(albumPaths))[:len(albumPaths)*7-1]
	if err := db.Where("album.owner_id = ? AND path_hash NOT IN ("+albums_questions+")", album_args...).Find(&albums).Error; err != nil {
		return []error{errors.Wrap(err, "get albums to be deleted from database")}
	}

	deleteErrors := make([]error, 0)

	deleted_album_ids := make([]interface{}, 0)
	for _, album := range albums {
		deleted_album_ids = append(deleted_album_ids, album.ID)
		cache_path := path.Join("./photo_cache", strconv.Itoa(int(album.ID)))
		err := os.RemoveAll(cache_path)
		if err != nil {
			deleteErrors = append(deleteErrors, errors.Wrapf(err, "delete unused cache folder (%s)", cache_path))
		}
	}

	if err := db.Delete(&albums).Error; err != nil {
		ScannerError("Could not delete old albums from database:\n%s\n", err)
		deleteErrors = append(deleteErrors, errors.Wrap(err, "delete old albums from database"))
	}

	return deleteErrors
}
