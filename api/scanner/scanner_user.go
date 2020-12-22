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

	if err := user.FillAlbums(db); err != nil {
		return nil, []error{err}
	}

	userAlbumIDs := make([]int, len(user.Albums))
	for i, album := range user.Albums {
		userAlbumIDs[i] = album.ID
	}

	var userRootAlbums []*models.Album
	if err := db.Where("id IN (?)", userAlbumIDs).Where("parent_album_id IS NULL").Find(&userRootAlbums).Error; err != nil {
		return nil, []error{err}
	}

	scanErrors := make([]error, 0)

	type scanInfo struct {
		path   string
		parent *models.Album
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
				path:   album.Path,
				parent: nil,
			})
		}
	}

	userAlbums := make([]*models.Album, 0)

	for scanQueue.Front() != nil {
		albumInfo := scanQueue.Front().Value.(scanInfo)
		scanQueue.Remove(scanQueue.Front())

		albumPath := albumInfo.path
		albumParent := albumInfo.parent

		// Read path
		dirContent, err := ioutil.ReadDir(albumPath)
		if err != nil {
			scanErrors = append(scanErrors, errors.Wrapf(err, "read directory (%s)", albumPath))
			continue
		}

		// Will become new album or album from db
		var album *models.Album

		transErr := db.Transaction(func(tx *gorm.DB) error {
			log.Printf("Scanning directory: %s", albumPath)

			// check if album already exists
			var albumResult []models.Album
			result := tx.Where("path_hash = md5(?)", albumPath).Find(&albumResult)
			if result.Error != nil {
				return result.Error
			}

			// album does not exist, create new
			if len(albumResult) == 0 {
				albumTitle := path.Base(albumPath)

				var albumParentID *int
				parentOwners := make([]models.User, 0)
				if albumParent != nil {
					albumParentID = &albumParent.ID

					if err := db.Model(&albumParent).Association("Owners").Find(&parentOwners); err != nil {
						return err
					}
				}

				album = &models.Album{
					Title:         albumTitle,
					ParentAlbumID: albumParentID,
					Path:          albumPath,
				}

				if err := tx.Create(&album).Error; err != nil {
					return errors.Wrap(err, "insert album into database")
				}

				if err := tx.Model(&album).Association("Owners").Append(parentOwners); err != nil {
					return errors.Wrap(err, "add owners to album")
				}
			} else {
				album = &albumResult[0]
			}

			userAlbums = append(userAlbums, album)

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
					path:   subalbumPath,
					parent: album,
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
