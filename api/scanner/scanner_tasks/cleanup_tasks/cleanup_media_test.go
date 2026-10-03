package cleanup_tasks_test

import (
	"os"
	"path"
	"testing"

	"github.com/otiai10/copy"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/face_detection"
	"github.com/photoview/photoview/api/scanner/scanner_tasks/cleanup_tasks"
	"github.com/photoview/photoview/api/test_utils"
	scanner_utils "github.com/photoview/photoview/api/test_utils/scanner"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	test_utils.IntegrationTestRun(m)
}

func TestCleanupMedia(t *testing.T) {
	test_utils.FilesystemTest(t)
	db := test_utils.DatabaseTest(t)

	if !assert.NoError(t, face_detection.InitializeFaceDetector(db)) {
		return
	}

	testDir := t.TempDir()
	assert.NoError(t, copy.Copy("../../test_media/library", testDir))

	countAllMedia := func() int {
		var allMedia []*models.Media
		if !assert.NoError(t, db.Find(&allMedia).Error) {
			return -1
		}
		return len(allMedia)
	}

	countAllMediaURLs := func() int {
		var allMediaURLs []*models.MediaURL
		if !assert.NoError(t, db.Find(&allMediaURLs).Error) {
			return -1
		}
		return len(allMediaURLs)
	}

	pass := "1234"
	user1, err := models.RegisterUser(db, "user1", &pass, true)
	if !assert.NoError(t, err) {
		return
	}

	user2, err := models.RegisterUser(db, "user2", &pass, true)
	if !assert.NoError(t, err) {
		return
	}

	rootAlbum := models.Album{
		Title: "root album",
		Path:  testDir,
	}

	if !assert.NoError(t, db.Save(&rootAlbum).Error) {
		return
	}

	err = db.Model(user1).Association("Albums").Append(&rootAlbum)
	if !assert.NoError(t, err) {
		return
	}
	err = db.Model(user2).Association("Albums").Append(&rootAlbum)
	if !assert.NoError(t, err) {
		return
	}

	t.Run("Modify albums", func(t *testing.T) {
		scanner_utils.RunScannerOnUser(t, db, user1)
		assert.Equal(t, 9, countAllMedia())
		assert.Equal(t, 18, countAllMediaURLs())

		// move faces directory
		assert.NoError(t, os.Rename(path.Join(testDir, "faces"), path.Join(testDir, "faces_moved")))
		scanner_utils.RunScannerAll(t, db)
		assert.Equal(t, 9, countAllMedia())
		assert.Equal(t, 18, countAllMediaURLs())

		// remove faces_moved directory
		assert.NoError(t, os.RemoveAll(path.Join(testDir, "faces_moved")))
		scanner_utils.RunScannerAll(t, db)
		assert.Equal(t, 3, countAllMedia())
		assert.Equal(t, 6, countAllMediaURLs())
	})

	t.Run("Modify images", func(t *testing.T) {
		assert.NoError(t, os.Rename(path.Join(testDir, "buttercup_close_summer_yellow.jpg"), path.Join(testDir, "yellow-flower.jpg")))
		scanner_utils.RunScannerAll(t, db)
		assert.Equal(t, 3, countAllMedia())
		assert.Equal(t, 6, countAllMediaURLs())

		assert.NoError(t, os.Remove(path.Join(testDir, "lilac_lilac_bush_lilac.jpg")))
		scanner_utils.RunScannerAll(t, db)
		assert.Equal(t, 2, countAllMedia())
		assert.Equal(t, 4, countAllMediaURLs())
	})
}

func TestDeleteOldUserAlbumsWithNothingLeftOnDisk(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	pass := "1234"
	user, err := models.RegisterUser(db, "empty-scan-user", &pass, true)
	if !assert.NoError(t, err) {
		return
	}

	album := models.Album{
		Title: "album that is gone",
		Path:  t.TempDir(),
	}
	if !assert.NoError(t, db.Save(&album).Error) {
		return
	}
	if !assert.NoError(t, db.Model(user).Association("Albums").Append(&album)) {
		return
	}

	// A complete walk that found no albums at all means every album the user
	// has is stale, so this must delete rather than bail out.
	errs := cleanup_tasks.DeleteOldUserAlbums(db, []*models.Album{}, user)
	assert.Empty(t, errs)

	var remaining int64
	assert.NoError(t, db.Model(&models.Album{}).Where("id = ?", album.ID).Count(&remaining).Error)
	assert.EqualValues(t, 0, remaining)
}
