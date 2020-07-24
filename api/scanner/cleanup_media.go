package scanner

import (
	"database/sql"
	"github.com/pkg/errors"
	"github.com/viktorstrate/photoview/api/graphql/models"
	"os"
	"path"
	"strconv"
	"strings"
)

func CleanupMedia(db *sql.DB, albumId int, albumPhotos []*models.Media) []error {
	if len(albumPhotos) == 0 {
		return nil
	}

	albumPhotoIds := make([]interface{}, len(albumPhotos))
	for i, photo := range albumPhotos {
		albumPhotoIds[i] = photo.MediaID
	}

	// Delete missing photos
	media_args := make([]interface{}, 0)
	media_args = append(media_args, albumId)
	media_args = append(media_args, albumPhotoIds...)

	media_questions := strings.Repeat("?,", len(albumPhotoIds))[:len(albumPhotoIds)*2-1]

	rows, err := db.Query(
		"SELECT media_id FROM media WHERE album_id = ? AND media_id NOT IN ("+media_questions+")",
		media_args...,
	)
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
		media_questions = strings.Repeat("?,", len(deleted_media_ids))[:len(deleted_media_ids)*2-1]

		if _, err := db.Exec("DELETE FROM media WHERE media_id IN ("+media_questions+")", deleted_media_ids...); err != nil {
			deleteErrors = append(deleteErrors, errors.Wrap(err, "delete old media from database"))
		}
	}

	return deleteErrors
}
