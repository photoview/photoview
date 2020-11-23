package scanner

import (
	"fmt"
	"io/ioutil"
	"path"
	"time"

	"github.com/pkg/errors"
	"github.com/viktorstrate/photoview/api/graphql/models"
	"github.com/viktorstrate/photoview/api/graphql/notification"
	"github.com/viktorstrate/photoview/api/utils"
	"gorm.io/gorm"
)

func scanAlbum(album *models.Album, cache *AlbumScannerCache, db *gorm.DB) {

	album_notify_key := utils.GenerateToken()
	notifyThrottle := utils.NewThrottle(500 * time.Millisecond)
	notifyThrottle.Trigger(nil)

	// Scan for photos
	albumPhotos, err := findMediaForAlbum(album, cache, db, func(photo *models.Media, newPhoto bool) {
		if newPhoto {
			notifyThrottle.Trigger(func() {
				notification.BroadcastNotification(&models.Notification{
					Key:     album_notify_key,
					Type:    models.NotificationTypeMessage,
					Header:  fmt.Sprintf("Found new media in album '%s'", album.Title),
					Content: fmt.Sprintf("Found %s", photo.Path),
				})
			})
		}
	})
	if err != nil {
		ScannerError("Failed to find media for album (%s): %s", album.Path, err)
	}

	album_has_changes := false
	for count, photo := range albumPhotos {
		// tx, err := db.Begin()

		transactionResult := db.Transaction(func(tx *gorm.DB) error {
			processing_was_needed, err := ProcessMedia(tx, photo)
			if err != nil {
				return errors.Wrapf(err, "failed to process photo (%s)", photo.Path)
			}

			if processing_was_needed {
				album_has_changes = true
				progress := float64(count) / float64(len(albumPhotos)) * 100.0
				notification.BroadcastNotification(&models.Notification{
					Key:      album_notify_key,
					Type:     models.NotificationTypeProgress,
					Header:   fmt.Sprintf("Processing media for album '%s'", album.Title),
					Content:  fmt.Sprintf("Processed media at %s", photo.Path),
					Progress: &progress,
				})
			}

			return nil
		})

		if transactionResult.Error != nil {
			ScannerError("Failed to begin database transaction: %s", transactionResult.Error)
		}
	}

	cleanup_errors := CleanupMedia(db, album.ID, albumPhotos)
	for _, err := range cleanup_errors {
		ScannerError("Failed to delete old media: %s", err)
	}

	if album_has_changes {
		timeoutDelay := 2000
		notification.BroadcastNotification(&models.Notification{
			Key:      album_notify_key,
			Type:     models.NotificationTypeMessage,
			Positive: true,
			Header:   fmt.Sprintf("Done processing media for album '%s'", album.Title),
			Content:  fmt.Sprintf("All media have been processed"),
			Timeout:  &timeoutDelay,
		})
	}
}

func findMediaForAlbum(album *models.Album, cache *AlbumScannerCache, db *gorm.DB, onScanPhoto func(photo *models.Media, newPhoto bool)) ([]*models.Media, error) {

	albumPhotos := make([]*models.Media, 0)

	dirContent, err := ioutil.ReadDir(album.Path)
	if err != nil {
		return nil, err
	}

	for _, item := range dirContent {
		photoPath := path.Join(album.Path, item.Name())

		if !item.IsDir() && isPathMedia(photoPath, cache) {

			db.Transaction(func(tx *gorm.DB) error {
				photo, isNewPhoto, err := ScanMedia(tx, photoPath, album.ID, cache)
				if err != nil {
					return errors.Wrapf(err, "Scanning media error (%s)", photoPath)
				}

				onScanPhoto(photo, isNewPhoto)

				albumPhotos = append(albumPhotos, photo)

				return nil
			})

		}
	}

	return albumPhotos, nil
}
