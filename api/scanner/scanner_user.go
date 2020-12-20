package scanner

import (
	"container/list"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/graphql/notification"
	"github.com/photoview/photoview/api/utils"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func findAlbumsForUser(db *gorm.DB, user *models.User, album_cache *AlbumScannerCache) ([]*models.Album, []error) {

	var userRootAlbums []*models.Album
	if err := db.Model(&user).Association("Albums").Find(&userRootAlbums); err != nil {
		return nil, []error{errors.Wrapf(err, "get albums of user (%s)", user.Username)}
	}

	scanErrors := make([]error, 0)

	type scanInfo struct {
		path     string
		parentID *int
	}

	scanQueue := list.New()

	for _, album := range userRootAlbums {
		// Check if user album directory exists on the file system
		if _, err := os.Stat(album.Path); err != nil {
			if os.IsNotExist(err) {
				scanErrors = append(scanErrors, errors.Errorf("Album directory for user '%s' does not exist '%s'\n", user.Username, album.Path))
			} else {
				scanErrors = append(scanErrors, errors.Errorf("Could not read album directory for user '%s': %s\n", user.Username, album.Path))
			}
		} else {
			scanQueue.PushBack(scanInfo{
				path:     album.Path,
				parentID: nil,
			})
		}
	}

	userAlbums := make([]*models.Album, 0)

	for scanQueue.Front() != nil {
		albumInfo := scanQueue.Front().Value.(scanInfo)
		scanQueue.Remove(scanQueue.Front())

		albumPath := albumInfo.path
		albumParentID := albumInfo.parentID

		// Read path
		dirContent, err := ioutil.ReadDir(albumPath)
		if err != nil {
			scanErrors = append(scanErrors, errors.Wrapf(err, "read directory (%s)", albumPath))
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
			scanErrors = append(scanErrors, errors.Wrap(transErr, "begin database transaction"))
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
	scanErrors = append(scanErrors, deleteErrors...)

	return userAlbums, scanErrors
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
