package resolvers

import (
	"os"
	"path"
	"strconv"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/face_detection"
	"github.com/photoview/photoview/api/utils"
	"gorm.io/gorm"
)

func cleanup(tx *gorm.DB, albumID int, childAlbumIDs []int) ([]int, error) {
	var userAlbumCount int
	var deletedAlbumIDs []int = nil

	if err := tx.Raw("SELECT COUNT(user_id) FROM user_albums WHERE album_id = ?",
		albumID).Scan(&userAlbumCount).Error; err != nil {

		return nil, err
	}

	if userAlbumCount == 0 {
		deletedAlbumIDs = append(childAlbumIDs, albumID)
		childAlbumIDs = nil
		// Delete albums from database
		if err := tx.Delete(&models.Album{}, "id IN (?)", deletedAlbumIDs).Error; err != nil {
			deletedAlbumIDs = nil
			return nil, err
		}
	}
	return deletedAlbumIDs, nil
}

func clearCacheAndReloadFaces(db *gorm.DB, deletedAlbumIDs []int) error {
	if deletedAlbumIDs != nil {
		// Delete albums from cache
		for _, id := range deletedAlbumIDs {
			cacheAlbumPath := path.Join(utils.MediaCachePath(), strconv.Itoa(id))

			if err := os.RemoveAll(cacheAlbumPath); err != nil {
				return err
			}
		}
		// Reload faces as media might have been deleted
		if face_detection.GlobalFaceDetector != nil {
			if err := face_detection.GlobalFaceDetector.ReloadFacesFromDatabase(db); err != nil {
				return err
			}
		}
	}
	return nil
}
