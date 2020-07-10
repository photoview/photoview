package scanner

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"path"
	"time"

	"github.com/viktorstrate/photoview/api/graphql/models"
	"github.com/viktorstrate/photoview/api/graphql/notification"
	"github.com/viktorstrate/photoview/api/utils"
)

func scanAlbum(album *models.Album, cache *AlbumScannerCache, db *sql.DB) {

	album_notify_key := utils.GenerateToken()
	notifyThrottle := utils.NewThrottle(500 * time.Millisecond)
	notifyThrottle.Trigger(nil)

	// Scan for photos
	albumPhotos, err := findPhotosForAlbum(album, cache, db, func(photo *models.Photo, newPhoto bool) {
		if newPhoto {
			notifyThrottle.Trigger(func() {
				notification.BroadcastNotification(&models.Notification{
					Key:     album_notify_key,
					Type:    models.NotificationTypeMessage,
					Header:  fmt.Sprintf("Found new photos in album '%s'", album.Title),
					Content: fmt.Sprintf("Found photo %s", photo.Path),
				})
			})
		}
	})
	if err != nil {
		ScannerError("Failed to find photos for album (%s): %s", album.Path, err)
	}

	album_has_changes := false
	for count, photo := range albumPhotos {
		tx, err := db.Begin()
		if err != nil {
			ScannerError("Failed to begin database transaction: %s", err)
		}

		processing_was_needed, err := ProcessMedia(tx, photo)
		if err != nil {
			tx.Rollback()
			ScannerError("Failed to process photo (%s): %s", photo.Path, err)
			continue
		}

		if processing_was_needed {
			album_has_changes = true
			progress := float64(count) / float64(len(albumPhotos)) * 100.0
			notification.BroadcastNotification(&models.Notification{
				Key:      album_notify_key,
				Type:     models.NotificationTypeProgress,
				Header:   fmt.Sprintf("Processing photo for album '%s'", album.Title),
				Content:  fmt.Sprintf("Processed photo at %s", photo.Path),
				Progress: &progress,
			})
		}

		err = tx.Commit()
		if err != nil {
			ScannerError("Failed to commit database transaction: %s", err)
		}
	}

	if album_has_changes {
		timeoutDelay := 2000
		notification.BroadcastNotification(&models.Notification{
			Key:      album_notify_key,
			Type:     models.NotificationTypeMessage,
			Positive: true,
			Header:   fmt.Sprintf("Done processing photos for album '%s'", album.Title),
			Content:  fmt.Sprintf("All photos have been processed"),
			Timeout:  &timeoutDelay,
		})
	}
}

func findPhotosForAlbum(album *models.Album, cache *AlbumScannerCache, db *sql.DB, onScanPhoto func(photo *models.Photo, newPhoto bool)) ([]*models.Photo, error) {

	albumPhotos := make([]*models.Photo, 0)

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
				ScannerError("Scanning image error (%s): %s", photoPath, err)
				tx.Rollback()
				continue
			}

			onScanPhoto(photo, isNewPhoto)

			albumPhotos = append(albumPhotos, photo)

			tx.Commit()
		}
	}

	return albumPhotos, nil
}
