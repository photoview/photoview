package scanner

import (
	"database/sql"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/viktorstrate/photoview/api/graphql/models"
)

func CleanupMedia(db *sql.DB, albumId int, albumMedia []*models.Media) []error {
	albumMediaIds := make([]interface{}, len(albumMedia))
	for i, photo := range albumMedia {
		albumMediaIds[i] = photo.MediaID
	}

	// Delete missing media
	var rows *sql.Rows
	var err error

	// Select media from database that was not found on hard disk
	if len(albumMedia) > 0 {
		media_args := make([]interface{}, 0)
		media_args = append(media_args, albumId)
		media_args = append(media_args, albumMediaIds...)

		media_questions := strings.Repeat("?,", len(albumMediaIds))[:len(albumMediaIds)*2-1]
		rows, err = db.Query(
			"SELECT media_id FROM media WHERE album_id = ? AND media_id NOT IN ("+media_questions+")",
			media_args...,
		)
	} else {
		rows, err = db.Query(
			"SELECT media_id FROM media WHERE album_id = ?",
			albumId,
		)
	}
	if err != nil {
		return []error{errors.Wrap(err, "get media files to be deleted from database")}
	}
	defer rows.Close()

	deleteErrors := make([]error, 0)

	deleted_media_ids := make([]interface{}, 0)
	for rows.Next() {
		var media_id int
		if err := rows.Scan(&media_id); err != nil {
			deleteErrors = append(deleteErrors, errors.Wrapf(err, "parse media to be removed (media_id %d)", media_id))
			continue
		}

		deleted_media_ids = append(deleted_media_ids, media_id)
		cache_path := path.Join(PhotoCache(), strconv.Itoa(albumId), strconv.Itoa(media_id))
		err := os.RemoveAll(cache_path)
		if err != nil {
			deleteErrors = append(deleteErrors, errors.Wrapf(err, "delete unused cache folder (%s)", cache_path))
		}
	}

	if len(deleted_media_ids) > 0 {
		media_questions := strings.Repeat("?,", len(deleted_media_ids))[:len(deleted_media_ids)*2-1]

		if _, err := db.Exec("DELETE FROM media WHERE media_id IN ("+media_questions+")", deleted_media_ids...); err != nil {
			deleteErrors = append(deleteErrors, errors.Wrap(err, "delete old media from database"))
		}
	}

	return deleteErrors
}

func deleteOldUserAlbums(db *sql.DB, scannedAlbums []*models.Album, user *models.User) []error {
	if len(scannedAlbums) == 0 {
		return nil
	}

	albumPaths := make([]interface{}, len(scannedAlbums))
	for i, album := range scannedAlbums {
		albumPaths[i] = album.Path
	}

	// Delete old albums
	album_args := make([]interface{}, 0)
	album_args = append(album_args, user.UserID)
	album_args = append(album_args, albumPaths...)

	albums_questions := strings.Repeat("MD5(?),", len(albumPaths))[:len(albumPaths)*7-1]
	rows, err := db.Query("SELECT album_id FROM album WHERE album.owner_id = ? AND path_hash NOT IN ("+albums_questions+")", album_args...)
	if err != nil {
		return []error{errors.Wrap(err, "get albums to be deleted from database")}
	}
	defer rows.Close()

	deleteErrors := make([]error, 0)

	deleted_album_ids := make([]interface{}, 0)
	for rows.Next() {
		var album_id int
		if err := rows.Scan(&album_id); err != nil {
			deleteErrors = append(deleteErrors, errors.Wrapf(err, "parse album to be removed (album_id %d)", album_id))
			continue
		}

		deleted_album_ids = append(deleted_album_ids, album_id)
		cache_path := path.Join("./photo_cache", strconv.Itoa(album_id))
		err := os.RemoveAll(cache_path)
		if err != nil {
			deleteErrors = append(deleteErrors, errors.Wrapf(err, "delete unused cache folder (%s)", cache_path))
		}
	}

	if len(deleted_album_ids) > 0 {
		albums_questions = strings.Repeat("?,", len(deleted_album_ids))[:len(deleted_album_ids)*2-1]

		if _, err := db.Exec("DELETE FROM album WHERE album_id IN ("+albums_questions+")", deleted_album_ids...); err != nil {
			ScannerError("Could not delete old albums from database:\n%s\n", err)
			deleteErrors = append(deleteErrors, errors.Wrap(err, "delete old albums from database"))
		}
	}

	return deleteErrors
}
