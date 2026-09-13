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
// scanner's media-type detection sniffs file content, not just the extension,
// so an arbitrary non-empty file isn't recognized as a photo and its directory
// never becomes an album.
func writeDummyPhoto(t *testing.T, dir, name string) {
	t.Helper()
	content, err := os.ReadFile("./test_media/orient/up_arrow_90cw_web.jpg")
	assert.NoError(t, err)
	assert.NoError(t, os.WriteFile(filepath.Join(dir, name), content, 0o644))
}

// TestFindAlbumsForAlbumRemovesVanishedSubAlbums covers moving a directory out
// of an album on disk: a rescan of that one album has to drop the album row the
// directory left behind, while leaving everything outside the scanned subtree
// alone.
func TestFindAlbumsForAlbumRemovesVanishedSubAlbums(t *testing.T) {
	db := test_utils.DatabaseTest(t)
	test_utils.FilesystemTest(t)

	rootPath := t.TempDir()
	root := models.Album{Title: "root", Path: rootPath}
	assert.NoError(t, db.Save(&root).Error)

	movedPath := filepath.Join(rootPath, "moved")
	assert.NoError(t, os.Mkdir(movedPath, 0o755))
	writeDummyPhoto(t, movedPath, "photo.jpg")

	stayingPath := filepath.Join(rootPath, "staying")
	assert.NoError(t, os.Mkdir(stayingPath, 0o755))
	writeDummyPhoto(t, stayingPath, "photo.jpg")

	// Never part of the scanned subtree, so the rescan never looks at its
	// directory and must not draw any conclusion about it.
	outsider := models.Album{Title: "outsider", Path: t.TempDir()}
	assert.NoError(t, db.Save(&outsider).Error)

	_, scanErrors := scanner.FindAlbumsForAlbum(db, &root, scanner_cache.MakeAlbumCache())
	assert.Empty(t, scanErrors)

	var moved models.Album
	assert.NoError(t, db.Where("path = ?", movedPath).First(&moved).Error)

	// The move itself - all this album's directory knows is that it's gone.
	assert.NoError(t, os.RemoveAll(movedPath))

	_, scanErrors = scanner.FindAlbumsForAlbum(db, &root, scanner_cache.MakeAlbumCache())
	assert.Empty(t, scanErrors)

	var count int64
	assert.NoError(t, db.Model(&models.Album{}).Where("id = ?", moved.ID).Count(&count).Error)
	assert.Zero(t, count, "a sub-album whose directory is gone must not survive a rescan of its parent")

	assert.NoError(t, db.Model(&models.Album{}).Where("path = ?", stayingPath).Count(&count).Error)
	assert.Equal(t, int64(1), count, "the sub-album still on disk must be kept")

	assert.NoError(t, db.Model(&models.Album{}).Where("id = ?", root.ID).Count(&count).Error)
	assert.Equal(t, int64(1), count, "the rescanned album itself must never be deleted")

	assert.NoError(t, db.Model(&models.Album{}).Where("id = ?", outsider.ID).Count(&count).Error)
	assert.Equal(t, int64(1), count, "an album outside the scanned subtree must be left alone")
}

// TestFindAlbumsForUserKeepsAlbumsWhenDiscoveryFails covers a root directory
// that can't be read - an unmounted share, say. Everything below it is missing
// from the walk's result, which the cleanup would otherwise read as "deleted".
func TestFindAlbumsForUserKeepsAlbumsWhenDiscoveryFails(t *testing.T) {
	db := test_utils.DatabaseTest(t)
	test_utils.FilesystemTest(t)

	user, err := models.RegisterUser(db, "incomplete_scan_user", nil, false)
	assert.NoError(t, err)

	readablePath := t.TempDir()
	readable := models.Album{Title: "readable", Path: readablePath}
	assert.NoError(t, db.Save(&readable).Error)
	writeDummyPhoto(t, readablePath, "photo.jpg")

	unreadablePath := t.TempDir()
	unreadable := models.Album{Title: "unreadable", Path: unreadablePath}
	assert.NoError(t, db.Save(&unreadable).Error)

	assert.NoError(t, db.Model(&user).Association("Albums").Append(&readable, &unreadable))

	// Not IsNotExist: the directory is there, we just can't look inside it.
	assert.NoError(t, os.Chmod(unreadablePath, 0o000))
	t.Cleanup(func() { os.Chmod(unreadablePath, 0o755) })

	// Permissions don't apply to a root-capable process, which is how tests
	// often run in a container - there the scan would succeed and there would
	// be no incomplete discovery to assert about.
	if _, err := os.ReadDir(unreadablePath); err == nil {
		t.Skip("this process can read a 0o000 directory, so discovery cannot fail here")
	}

	_, scanErrors := scanner.FindAlbumsForUser(db, user, scanner_cache.MakeAlbumCache())
	assert.NotEmpty(t, scanErrors, "the unreadable directory should be reported")

	var count int64
	assert.NoError(t, db.Model(&models.Album{}).Where("id = ?", unreadable.ID).Count(&count).Error)
	assert.Equal(t, int64(1), count, "an album must not be deleted just because the scan couldn't read it")
}
