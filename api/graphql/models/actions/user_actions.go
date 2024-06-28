package actions

import (
	"errors"
	"github.com/photoview/photoview/api/database/drivers"
	"os"
	"path"
	"strconv"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/utils"
	"gorm.io/gorm"
)

func DeleteUser(db *gorm.DB, userID int) (*models.User, error) {

	// make sure the last admin user is not deleted
	var adminUsers []*models.User
	var query string
	if drivers.POSTGRES.MatchDatabase(db) {
		query = "\"Role\".name = 'ADMIN'"
	} else {
		query = "Role.Name = 'ADMIN'"
	}
	db.Debug().Joins("Role").Model(&models.User{}).
		Where(query).
		Limit(2).
		Find(&adminUsers)

	if len(adminUsers) == 1 && adminUsers[0].ID == userID {
		return nil, errors.New("deleting sole admin user is not allowed")
	}

	var user models.User
	deletedAlbumIDs := make([]int, 0)

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}

		userAlbums := user.Albums
		if err := tx.Model(&user).Association("Albums").Find(&userAlbums); err != nil {
			return err
		}

		if err := tx.Model(&user).Association("Albums").Clear(); err != nil {
			return err
		}

		for _, album := range userAlbums {
			var associatedUsers = tx.Model(album).Association("Owners").Count()

			if associatedUsers == 0 {
				deletedAlbumIDs = append(deletedAlbumIDs, album.ID)
				if err := tx.Delete(album).Error; err != nil {
					return err
				}
			}
		}

		if err := tx.Delete(&user).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// If there is only one associated user, clean up the cache folder and delete the album row
	for _, deletedAlbumID := range deletedAlbumIDs {
		cachePath := path.Join(utils.MediaCachePath(), strconv.Itoa(int(deletedAlbumID)))
		if err := os.RemoveAll(cachePath); err != nil {
			return &user, err
		}
	}
	return &user, nil
}
