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
		assert.Equal(t, "./test_data", album.Path)
		assert.NotEmpty(t, album.Owners)
	})

	t.Run("Insert duplicate root album", func(t *testing.T) {

		_, err := scanner.NewRootAlbum(db, "./test_data", &user)

		assert.Error(t, err)
		assert.Equal(t, err.Error(), "user already owns path (./test_data)")
	})

}
