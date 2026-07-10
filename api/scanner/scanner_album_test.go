package scanner_test

import (
	"os"
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
)

const testDataPath = "./test_media/library"

func TestNewRootPath(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	user := models.User{
		Username: "user1",
	}

	if !assert.NoError(t, db.Save(&user).Error) {
		return
	}

	t.Run("Insert valid root album", func(t *testing.T) {
		album, err := scanner.NewRootAlbum(db, testDataPath, &user)
		if !assert.NoError(t, err) {
			return
		}

		assert.NotNil(t, album)
		assert.Contains(t, album.Path, "/api/scanner/test_media")
		assert.NotEmpty(t, album.Owners)
	})

	t.Run("Insert duplicate root album", func(t *testing.T) {

		_, err := scanner.NewRootAlbum(db, testDataPath, &user)

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

		album, err := scanner.NewRootAlbum(db, testDataPath, &user2)
		if !assert.NoError(t, err) {
			return
		}

		assert.NotNil(t, album)
		assert.Contains(t, album.Path, "/api/scanner/test_media")

		ownerCount := db.Model(&album).Association("Owners").Count()
		assert.EqualValues(t, 2, ownerCount)
	})

	t.Run("Insert root album pointing to a file", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "photoview-rootpath-test-*")
		if !assert.NoError(t, err) {
			return
		}
		tmpFile.Close()
		defer os.Remove(tmpFile.Name())

		_, err = scanner.NewRootAlbum(db, tmpFile.Name(), &user)
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "invalid root path")
	})
}

func TestValidRootPath(t *testing.T) {
	t.Run("Valid directory", func(t *testing.T) {
		assert.True(t, scanner.ValidRootPath(testDataPath))
	})

	t.Run("Non-existent path", func(t *testing.T) {
		assert.False(t, scanner.ValidRootPath("./non_existent_path_xyz"))
	})

	t.Run("File path is rejected", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "photoview-validroot-test-*")
		if !assert.NoError(t, err) {
			return
		}
		tmpFile.Close()
		defer os.Remove(tmpFile.Name())

		assert.False(t, scanner.ValidRootPath(tmpFile.Name()))
	})

	t.Run("Path with dot-slash prefix resolves correctly", func(t *testing.T) {
		// A path like "./test_media/library/./" should resolve to the same
		// real directory and be accepted.
		assert.True(t, scanner.ValidRootPath(testDataPath+"/./"))
	})
}
