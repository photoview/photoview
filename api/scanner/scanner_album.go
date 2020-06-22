package scanner

import (
	"database/sql"
	"io/ioutil"
	"path"

	"github.com/viktorstrate/photoview/api/graphql/models"
)

func findPhotosForAlbum(album *models.Album, cache *ScannerCache, db *sql.DB, onScanPhoto func(photo *models.Photo, newPhoto bool)) ([]*models.Photo, error) {

	newPhotos := make([]*models.Photo, 0)

	dirContent, err := ioutil.ReadDir(album.Path)
	if err != nil {
		return nil, err
	}

	for _, item := range dirContent {
		photoPath := path.Join(album.Path, item.Name())

		if !item.IsDir() && isPathImage(photoPath, cache) {
			tx, err := db.Begin()
			if err != nil {
				ScannerError("Could not begin database transaction for image %s: %s\n", photoPath, err)
				continue
			}

			cache.photo_paths_scanned = append(cache.photo_paths_scanned, photoPath)

			photo, isNewPhoto, err := ScanPhoto(tx, photoPath, album.AlbumID)
			if err != nil {
				ScannerError("Scanning image %s: %s", photoPath, err)
				tx.Rollback()
				continue
			}

			onScanPhoto(photo, isNewPhoto)

			if isNewPhoto {
				newPhotos = append(newPhotos, photo)
			}

			tx.Commit()
		}
	}

	return newPhotos, nil
}
