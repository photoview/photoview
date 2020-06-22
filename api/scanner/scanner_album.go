package scanner

import (
	"database/sql"
	"io/ioutil"
	"path"

	"github.com/viktorstrate/photoview/api/graphql/models"
)

func scanAlbum(album *models.Album, cache *AlbumScannerCache, db *sql.DB) {
	// Scan for photos
	albumPhotos, err := findPhotosForAlbum(album, cache, db, func(photo *models.Photo, newPhoto bool) {
		// notifyThrottle.Trigger(func() {
		// 	notification.BroadcastNotification(&models.Notification{
		// 		Key:     processKey,
		// 		Type:    models.NotificationTypeMessage,
		// 		Header:  fmt.Sprintf("Scanning photo for user '%s'", user.Username),
		// 		Content: fmt.Sprintf("Scanning image at %s", photo.Path),
		// 	})
		// })
	})
	if err != nil {
		ScannerError("Failed to find photos for album (%s): %s", album.Path, err)
	}

	tx, err := db.Begin()
	if err != nil {
		ScannerError("Failed to begin database transaction: %s", err)
	}

	for _, photo := range albumPhotos {
		err = ProcessPhoto(tx, photo)
		if err != nil {
			ScannerError("Failed to process photo (%s): %s", photo.Path, err)
		}

		// TODO: Broadcast progress
	}
}

func findPhotosForAlbum(album *models.Album, cache *AlbumScannerCache, db *sql.DB, onScanPhoto func(photo *models.Photo, newPhoto bool)) ([]*models.Photo, error) {

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
