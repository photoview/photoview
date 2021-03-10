package scanner

import (
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/graphql/notification"
	"github.com/photoview/photoview/api/scanner/face_detection"
	"github.com/photoview/photoview/api/utils"
	"github.com/pkg/errors"
	"github.com/sabhiram/go-gitignore"
	"gorm.io/gorm"
)

func NewRootAlbum(db *gorm.DB, rootPath string, owner *models.User) (*models.Album, error) {

	owners := []models.User{
		*owner,
	}

	var matchedAlbums []models.Album
	if err := db.Where("path_hash = ?", models.MD5Hash(rootPath)).Find(&matchedAlbums).Error; err != nil {
		return nil, err
	}

	if len(matchedAlbums) > 0 {
		album := matchedAlbums[0]

		if err := db.Model(&owner).Association("Albums").Append(&album); err != nil {
			return nil, errors.Wrap(err, "failed to add owner to already existing album")
		}

		return &album, nil
	} else {
		album := models.Album{
			Title:  path.Base(rootPath),
			Path:   rootPath,
			Owners: owners,
		}

		if err := db.Create(&album).Error; err != nil {
			return nil, err
		}

		return &album, nil
	}
}

func scanAlbum(album *models.Album, cache *AlbumScannerCache, db *gorm.DB) {

	album_notify_key := utils.GenerateToken()
	notifyThrottle := utils.NewThrottle(500 * time.Millisecond)
	notifyThrottle.Trigger(nil)

	// Scan for photos
	albumMedia, err := findMediaForAlbum(album, cache, db, func(photo *models.Media, newPhoto bool) {
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
	for count, media := range albumMedia {
		// tx, err := db.Begin()

		transactionError := db.Transaction(func(tx *gorm.DB) error {
			processing_was_needed, err := ProcessMedia(tx, media)
			if err != nil {
				return errors.Wrapf(err, "failed to process photo (%s)", media.Path)
			}

			if processing_was_needed {
				album_has_changes = true
				progress := float64(count) / float64(len(albumMedia)) * 100.0
				notification.BroadcastNotification(&models.Notification{
					Key:      album_notify_key,
					Type:     models.NotificationTypeProgress,
					Header:   fmt.Sprintf("Processing media for album '%s'", album.Title),
					Content:  fmt.Sprintf("Processed media at %s", media.Path),
					Progress: &progress,
				})

				if media.Type == models.MediaTypePhoto {
					go func() {
						if err := face_detection.GlobalFaceDetector.DetectFaces(media); err != nil {
							ScannerError("Error detecting faces in image (%s): %s", media.Path, err)
						}
					}()
				}
			}

			return nil
		})

		if transactionError != nil {
			ScannerError("Failed to begin database transaction: %s", transactionError)
		}
	}

	cleanup_errors := CleanupMedia(db, album.ID, albumMedia)
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

	// Get ignore data
	 albumIgnore := ignore.CompileIgnoreLines(*cache.GetAlbumIgnore(album.Path)...)

	for _, item := range dirContent {
		photoPath := path.Join(album.Path, item.Name())

		if !item.IsDir() && isPathMedia(photoPath, cache) {
			// Match file against ignore data
			if (albumIgnore.MatchesPath(item.Name())) {
				log.Printf("File %s ignored\n", item.Name())
				continue
			}

			// Skip the JPEGs that are compressed version of raw files
			counterpartFile := scanForRawCounterpartFile(photoPath)
			if counterpartFile != nil {
				continue
			}

			err := db.Transaction(func(tx *gorm.DB) error {
				media, isNewMedia, err := ScanMedia(tx, photoPath, album.ID, cache)
				if err != nil {
					return errors.Wrapf(err, "Scanning media error (%s)", photoPath)
				}

				onScanPhoto(media, isNewMedia)

				albumPhotos = append(albumPhotos, media)

				return nil
			})

			if err != nil {
				ScannerError("Error scanning media for album (%d): %s\n", album.ID, err)
				continue
			}

		}
	}

	return albumPhotos, nil
}
