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

func CleanupMedia(db *gorm.DB, albumId int, albumMedia []*models.Media) []error {
	albumMediaIds := make([]int, len(albumMedia))
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

	mediaIDs := make([]int, 0)
	for _, media := range mediaList {

		mediaIDs = append(mediaIDs, media.ID)
		cachePath := path.Join(PhotoCache(), strconv.Itoa(int(albumId)), strconv.Itoa(int(media.ID)))
		err := os.RemoveAll(cachePath)
		if err != nil {
			deleteErrors = append(deleteErrors, errors.Wrapf(err, "delete unused cache folder (%s)", cachePath))
		}

	}

	if err := db.Where("id IN ?", mediaIDs).Delete(models.Media{}).Error; err != nil {
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
	if err := db.Where("owner_id = ? AND path_hash NOT IN ("+albums_questions+")", album_args...).Find(&albums).Error; err != nil {
		return []error{errors.Wrap(err, "get albums to be deleted from database")}
	}

	deleteErrors := make([]error, 0)

	albumIDs := make([]int, 0)
	for _, album := range albums {
		albumIDs = append(albumIDs, album.ID)
		cachePath := path.Join("./photo_cache", strconv.Itoa(int(album.ID)))
		err := os.RemoveAll(cachePath)
		if err != nil {
			deleteErrors = append(deleteErrors, errors.Wrapf(err, "delete unused cache folder (%s)", cachePath))
		}
	}

	if err := db.Where("id IN ?", albumIDs).Delete(models.Album{}).Error; err != nil {
		ScannerError("Could not delete old albums from database:\n%s\n", err)
		deleteErrors = append(deleteErrors, errors.Wrap(err, "delete old albums from database"))
	}

	return deleteErrors
}
