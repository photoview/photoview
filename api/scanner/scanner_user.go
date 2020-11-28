package scanner

import (
	"container/list"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/viktorstrate/photoview/api/graphql/models"
	"github.com/viktorstrate/photoview/api/graphql/notification"
	"github.com/viktorstrate/photoview/api/utils"
	"gorm.io/gorm"
)

func findAlbumsForUser(db *gorm.DB, user *models.User, album_cache *AlbumScannerCache) ([]*models.Album, []error) {

	// Check if user directory exists on the file system
	if _, err := os.Stat(user.RootPath); err != nil {
		if os.IsNotExist(err) {
			return nil, []error{errors.Errorf("Photo directory for user '%s' does not exist '%s'\n", user.Username, user.RootPath)}
		} else {
			return nil, []error{errors.Errorf("Could not read photo directory for user '%s': %s\n", user.Username, user.RootPath)}
		}
	}

	type scanInfo struct {
		path     string
		parentID *int
	}

	scanQueue := list.New()
	scanQueue.PushBack(scanInfo{
		path:     user.RootPath,
		parentID: nil,
	})

	userAlbums := make([]*models.Album, 0)
	albumErrors := make([]error, 0)
	// newPhotos := make([]*models.Photo, 0)

	for scanQueue.Front() != nil {
		albumInfo := scanQueue.Front().Value.(scanInfo)
		scanQueue.Remove(scanQueue.Front())

		albumPath := albumInfo.path
		albumParentID := albumInfo.parentID

		// Read path
		dirContent, err := ioutil.ReadDir(albumPath)
		if err != nil {
			albumErrors = append(albumErrors, errors.Wrapf(err, "read directory (%s)", albumPath))
			continue
		}

		// Will become new album or album from db
		var album models.Album

		transErr := db.Transaction(func(tx *gorm.DB) error {
			log.Printf("Scanning directory: %s", albumPath)

			// Make album if not exists
			albumTitle := path.Base(albumPath)

			err = tx.FirstOrCreate(&album, models.Album{
				Title:         albumTitle,
				ParentAlbumID: albumParentID,
				OwnerID:       user.ID,
				Path:          albumPath,
			}).Error

			if err != nil {
				return errors.Wrap(err, "insert album into database")
			}

			userAlbums = append(userAlbums, &album)

			return nil
		})

		if transErr != nil {
			albumErrors = append(albumErrors, errors.Wrap(transErr, "begin database transaction"))
			continue
		}

		// Scan for sub-albums
		for _, item := range dirContent {
			subalbumPath := path.Join(albumPath, item.Name())

			// Skip if directory is hidden
			if path.Base(subalbumPath)[0:1] == "." {
				continue
			}

			if item.IsDir() && directoryContainsPhotos(subalbumPath, album_cache) {
				scanQueue.PushBack(scanInfo{
					path:     subalbumPath,
					parentID: &album.ID,
				})
			}
		}
	}

	deleteErrors := deleteOldUserAlbums(db, userAlbums, user)
	albumErrors = append(albumErrors, deleteErrors...)

	return userAlbums, albumErrors
}

func directoryContainsPhotos(rootPath string, cache *AlbumScannerCache) bool {

	if contains_image := cache.AlbumContainsPhotos(rootPath); contains_image != nil {
		return *contains_image
	}

	scanQueue := list.New()
	scanQueue.PushBack(rootPath)

	scanned_directories := make([]string, 0)

	for scanQueue.Front() != nil {

		dirPath := scanQueue.Front().Value.(string)
		scanQueue.Remove(scanQueue.Front())

		scanned_directories = append(scanned_directories, dirPath)

		dirContent, err := ioutil.ReadDir(dirPath)
		if err != nil {
			ScannerError("Could not read directory: %s\n", err.Error())
			return false
		}

		for _, fileInfo := range dirContent {
			filePath := path.Join(dirPath, fileInfo.Name())
			if fileInfo.IsDir() {
				scanQueue.PushBack(filePath)
			} else {
				if isPathMedia(filePath, cache) {
					cache.InsertAlbumPaths(dirPath, rootPath, true)
					return true
				}
			}
		}

	}

	for _, scanned_path := range scanned_directories {
		cache.InsertAlbumPath(scanned_path, false)
	}
	return false
}

func ScannerError(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)

	log.Printf("ERROR: %s", message)
	notification.BroadcastNotification(&models.Notification{
		Key:      utils.GenerateToken(),
		Type:     models.NotificationTypeMessage,
		Header:   "Scanner error",
		Content:  message,
		Negative: true,
	})
}

func PhotoCache() string {
	photoCache := os.Getenv("PHOTO_CACHE")
	if photoCache == "" {
		photoCache = "./photo_cache"
	}

	return photoCache
}
