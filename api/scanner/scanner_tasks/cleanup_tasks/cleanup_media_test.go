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

	// PropagateAlbumLevel (not a raw many2many Association.Append) so
	// ownership is backed by a real UserAlbumGrant row, matching how
	// NewRootAlbum grants a real root path - the scanner's own
	// grant-copying for newly discovered sub-albums reads from
	// UserAlbumGrant, not the materialized UserAlbums row directly.
	if !assert.NoError(t, models.PropagateAlbumLevel(db, rootAlbum.ID, user1.ID, models.AlbumPermissionLevelRead, nil)) {
		return
	}
	if !assert.NoError(t, models.PropagateAlbumLevel(db, rootAlbum.ID, user2.ID, models.AlbumPermissionLevelRead, nil)) {
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

// TestDeleteOldUserAlbumsCleansUpStrandedGrantSources covers a real bug: a
// deleted album can still be referenced as a grant's SourceAlbumID by a
// surviving descendant a few levels down (not itself stale, so it isn't
// among the deleted albums). UserAlbumGrant.SourceAlbumID has no actual
// foreign key (GORM doesn't create one for a scalar field without an
// association), so nothing enforces that only cleaning up rows scoped to
// AlbumID also catches rows that merely cite a deleted album as their
// source - left behind, the survivor's materialized UserAlbums row would
// stay stuck at whatever level included that now-gone source.
func TestDeleteOldUserAlbumsCleansUpStrandedGrantSources(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	user, err := models.RegisterUser(db, "stranded_grant_user", nil, false)
	assert.NoError(t, err)

	staleAlbum := models.Album{Title: "stale", Path: t.TempDir()}
	assert.NoError(t, db.Save(&staleAlbum).Error)
	survivor := models.Album{Title: "survivor", Path: t.TempDir(), ParentAlbumID: &staleAlbum.ID}
	assert.NoError(t, db.Save(&survivor).Error)

	// staleAlbum's own grant - covered by the existing AlbumID-scoped
	// cleanup, included here so its removal doesn't mask the SourceAlbumID
	// gap being tested.
	assert.NoError(t, db.Create(&models.UserAlbumGrant{
		UserID: user.ID, AlbumID: staleAlbum.ID, SourceAlbumID: staleAlbum.ID, Level: models.AlbumPermissionLevelRead,
	}).Error)
	assert.NoError(t, db.Create(&models.UserAlbums{UserID: user.ID, AlbumID: staleAlbum.ID, Level: models.AlbumPermissionLevelRead}).Error)

	// survivor has two independent sources: one from staleAlbum (about to be
	// stranded) at the higher level, and one self-sourced that must remain
	// untouched.
	assert.NoError(t, db.Create(&models.UserAlbumGrant{
		UserID: user.ID, AlbumID: survivor.ID, SourceAlbumID: staleAlbum.ID, Level: models.AlbumPermissionLevelUpload,
	}).Error)
	assert.NoError(t, db.Create(&models.UserAlbumGrant{
		UserID: user.ID, AlbumID: survivor.ID, SourceAlbumID: survivor.ID, Level: models.AlbumPermissionLevelRead,
	}).Error)
	assert.NoError(t, db.Create(&models.UserAlbums{UserID: user.ID, AlbumID: survivor.ID, Level: models.AlbumPermissionLevelUpload}).Error)

	// Only survivor was found on this scan - staleAlbum is gone from disk.
	deleteErrors := cleanup_tasks.DeleteOldUserAlbums(db, []*models.Album{&survivor}, user)
	assert.Empty(t, deleteErrors)

	var staleAlbumCount int64
	assert.NoError(t, db.Model(&models.Album{}).Where("id = ?", staleAlbum.ID).Count(&staleAlbumCount).Error)
	assert.Zero(t, staleAlbumCount, "the stale album itself should be deleted")

	var strandedGrantCount int64
	assert.NoError(t, db.Model(&models.UserAlbumGrant{}).
		Where("album_id = ? AND source_album_id = ?", survivor.ID, staleAlbum.ID).
		Count(&strandedGrantCount).Error)
	assert.Zero(t, strandedGrantCount, "the grant sourced from the deleted album must be cleaned up even though the target album survives")

	var survivorGrant models.UserAlbums
	assert.NoError(t, db.Where("user_id = ? AND album_id = ?", user.ID, survivor.ID).First(&survivorGrant).Error)
	assert.Equal(t, models.AlbumPermissionLevelRead, survivorGrant.Level, "the survivor's materialized level must be recomputed down to its one remaining source")
}
