package scanner_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner"
	"github.com/photoview/photoview/api/scanner/scanner_cache"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
)

// writeDummyPhoto copies a real (tiny) fixture image into dir/name - the
// scanner's media-type detection sniffs file content, not just extension,
// so an arbitrary non-empty file isn't recognized as a photo. Actual
// decoding/thumbnailing happens later in a separate scan job;
// FindAlbumsForAlbum/FindAlbumsForUser only need directoryContainsPhotos to
// recognize this as media so the containing directory gets discovered as an
// album.
func writeDummyPhoto(t *testing.T, dir, name string) {
	t.Helper()
	content, err := os.ReadFile("./test_media/orient/up_arrow_90cw_web.jpg")
	assert.NoError(t, err)
	assert.NoError(t, os.WriteFile(filepath.Join(dir, name), content, 0o644))
}

// TestFindAlbumsForAlbumPreservesGrantProvenance covers a real bug: when the
// scanner discovers a new sub-album on disk, it used to copy the parent's
// materialized UserAlbums row directly, instead of a UserAlbumGrant row
// carrying the grant's true SourceAlbumID. That meant a later revoke of the
// real source only still reached the new sub-album by coincidence, as long
// as it stayed under that source in the album tree - moving it out first
// (MoveAlbum) left the copied access stranded forever, an authorization
// bypass.
func TestFindAlbumsForAlbumPreservesGrantProvenance(t *testing.T) {
	db := test_utils.DatabaseTest(t)
	test_utils.FilesystemTest(t)

	user, err := models.RegisterUser(db, "grant_provenance_user", nil, false)
	assert.NoError(t, err)

	// A separate ancestor album, further up the tree than root - root's own
	// grant is sourced from here, not from root itself.
	ancestor := models.Album{Title: "ancestor", Path: t.TempDir()}
	assert.NoError(t, db.Save(&ancestor).Error)

	rootPath := t.TempDir()
	root := models.Album{Title: "root", Path: rootPath}
	assert.NoError(t, db.Save(&root).Error)

	assert.NoError(t, db.Create(&models.UserAlbumGrant{
		UserID: user.ID, AlbumID: root.ID, SourceAlbumID: ancestor.ID,
		Level: models.AlbumPermissionLevelUpload,
	}).Error)
	assert.NoError(t, db.Create(&models.UserAlbums{
		UserID: user.ID, AlbumID: root.ID, Level: models.AlbumPermissionLevelUpload,
	}).Error)

	childPath := filepath.Join(rootPath, "child")
	assert.NoError(t, os.Mkdir(childPath, 0o755))
	writeDummyPhoto(t, childPath, "photo.jpg")

	cache := scanner_cache.MakeAlbumCache()
	_, scanErrors := scanner.FindAlbumsForAlbum(db, &root, cache)
	assert.Empty(t, scanErrors)

	var child models.Album
	assert.NoError(t, db.Where("path = ?", childPath).First(&child).Error)

	var childGrant models.UserAlbumGrant
	err = db.Where("user_id = ? AND album_id = ?", user.ID, child.ID).First(&childGrant).Error
	if assert.NoError(t, err, "the new sub-album must get its own UserAlbumGrant row, not just a materialized copy") {
		assert.Equal(t, ancestor.ID, childGrant.SourceAlbumID, "the copied grant must keep the original source, not attribute it to root")
		assert.Equal(t, models.AlbumPermissionLevelUpload, childGrant.Level)
	}

	var childUserAlbums models.UserAlbums
	assert.NoError(t, db.Where("user_id = ? AND album_id = ?", user.ID, child.ID).First(&childUserAlbums).Error)
	assert.Equal(t, models.AlbumPermissionLevelUpload, childUserAlbums.Level)

	// The regression check: revoking the true source must reach the new
	// sub-album even after it's been moved out from under root, which only
	// works if the copied grant carries the real SourceAlbumID.
	elsewhere := models.Album{Title: "elsewhere", Path: t.TempDir()}
	assert.NoError(t, db.Save(&elsewhere).Error)
	child.ParentAlbumID = &elsewhere.ID
	assert.NoError(t, db.Save(&child).Error)

	assert.NoError(t, models.RevokeAlbumLevel(db, ancestor.ID, user.ID))

	var count int64
	assert.NoError(t, db.Model(&models.UserAlbums{}).
		Where("user_id = ? AND album_id = ?", user.ID, child.ID).Count(&count).Error)
	assert.Zero(t, count, "revoking the true source must remove access even after the sub-album moved elsewhere")
}

// TestFindAlbumsForUserPreservesGrantProvenanceOnSecondRootPath covers the
// second grant-copying path in the scanner: an album already exists (e.g.
// reachable through a second root path) and a particular user doesn't yet
// have their own row on it, so the scanner copies whatever they hold on its
// parent. That copy must also preserve the original SourceAlbumID.
func TestFindAlbumsForUserPreservesGrantProvenanceOnSecondRootPath(t *testing.T) {
	db := test_utils.DatabaseTest(t)
	test_utils.FilesystemTest(t)

	user, err := models.RegisterUser(db, "second_root_user", nil, false)
	assert.NoError(t, err)

	ancestor := models.Album{Title: "ancestor2", Path: t.TempDir()}
	assert.NoError(t, db.Save(&ancestor).Error)

	rootPath := t.TempDir()
	root := models.Album{Title: "root2", Path: rootPath}
	assert.NoError(t, db.Save(&root).Error)

	assert.NoError(t, db.Create(&models.UserAlbumGrant{
		UserID: user.ID, AlbumID: root.ID, SourceAlbumID: ancestor.ID,
		Level: models.AlbumPermissionLevelRead,
	}).Error)
	assert.NoError(t, db.Create(&models.UserAlbums{
		UserID: user.ID, AlbumID: root.ID, Level: models.AlbumPermissionLevelRead,
	}).Error)

	// The child album already exists in the DB (as if discovered through
	// another root path) but this user has no row on it yet.
	childPath := filepath.Join(rootPath, "child2")
	assert.NoError(t, os.Mkdir(childPath, 0o755))
	writeDummyPhoto(t, childPath, "photo.jpg")
	child := models.Album{Title: "child2", Path: childPath, ParentAlbumID: &root.ID}
	assert.NoError(t, db.Save(&child).Error)

	cache := scanner_cache.MakeAlbumCache()
	_, scanErrors := scanner.FindAlbumsForUser(db, user, cache)
	assert.Empty(t, scanErrors)

	var childGrant models.UserAlbumGrant
	err = db.Where("user_id = ? AND album_id = ?", user.ID, child.ID).First(&childGrant).Error
	if assert.NoError(t, err, "the user must get their own UserAlbumGrant row on the pre-existing child") {
		assert.Equal(t, ancestor.ID, childGrant.SourceAlbumID, "the copied grant must keep the original source, not attribute it to root")
		assert.Equal(t, models.AlbumPermissionLevelRead, childGrant.Level)
	}
}
