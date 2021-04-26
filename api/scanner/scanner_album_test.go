package scanner_test

import (
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
)

func TestNewRootPath(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	user := models.User{
		Username: "user1",
	}

	if !assert.NoError(t, db.Save(&user).Error) {
		return
	}

	t.Run("Insert valid root album", func(t *testing.T) {
		album, err := scanner.NewRootAlbum(db, "./test_data", &user)
		if !assert.NoError(t, err) {
			return
		}

		assert.NotNil(t, album)
		assert.Contains(t, album.Path, "/api/scanner/test_data")
		assert.NotEmpty(t, album.Owners)
	})

	t.Run("Insert duplicate root album", func(t *testing.T) {

		_, err := scanner.NewRootAlbum(db, "./test_data", &user)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user already owns a path containing this path:")
	})

	t.Run("Insert invalid root album", func(t *testing.T) {

		_, err := scanner.NewRootAlbum(db, "./invalid_path", &user)

		assert.Error(t, err)
		assert.Equal(t, err.Error(), "invalid root path")
	})

	t.Run("Add existing root album to new user", func(t *testing.T) {

		user2 := models.User{
			Username: "user2",
		}

		if !assert.NoError(t, db.Save(&user2).Error) {
			return
		}

		album, err := scanner.NewRootAlbum(db, "./test_data", &user2)
		if !assert.NoError(t, err) {
			return
		}

		assert.NotNil(t, album)
		assert.Contains(t, album.Path, "/api/scanner/test_data")

		owner_count := db.Model(&album).Association("Owners").Count()
		assert.EqualValues(t, 2, owner_count)
	})

}
