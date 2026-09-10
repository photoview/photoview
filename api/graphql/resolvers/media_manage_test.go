package resolvers

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
)

// setupMediaManageTest creates an album on a real temp directory, an
// uploader (Upload-level), a read-only user (Read-level) and an admin (no
// explicit grant needed), all sharing that one album.
func setupMediaManageTest(t *testing.T) (r *mutationResolver, album *models.Album, uploader *models.User, readOnly *models.User, admin *models.User) {
	t.Helper()

	test_utils.FilesystemTest(t)
	db := test_utils.DatabaseTest(t)

	albumPath := t.TempDir()
	album = &models.Album{Title: "album", Path: albumPath}
	assert.NoError(t, db.Save(album).Error)

	var err error
	uploader, err = models.RegisterUser(db, "media_uploader", nil, false)
	assert.NoError(t, err)
	readOnly, err = models.RegisterUser(db, "media_read_only", nil, false)
	assert.NoError(t, err)
	admin, err = models.RegisterUser(db, "media_admin", nil, true)
	assert.NoError(t, err)

	assert.NoError(t, db.Create(&models.UserAlbums{
		UserID: uploader.ID, AlbumID: album.ID, Level: models.AlbumPermissionLevelUpload,
	}).Error)
	assert.NoError(t, db.Create(&models.UserAlbums{
		UserID: readOnly.ID, AlbumID: album.ID, Level: models.AlbumPermissionLevelRead,
	}).Error)

	r = &mutationResolver{Resolver: &Resolver{database: db}}

	return r, album, uploader, readOnly, admin
}

func makeTestMediaFile(t *testing.T, r *mutationResolver, album *models.Album, fileName string) *models.Media {
	t.Helper()

	mediaPath := filepath.Join(album.Path, fileName)
	assert.NoError(t, os.WriteFile(mediaPath, []byte("fake media content"), 0o644))

	media := &models.Media{Title: fileName, Path: mediaPath, AlbumID: album.ID}
	assert.NoError(t, r.database.Save(media).Error)

	return media
}

func TestRenameMedia(t *testing.T) {
	r, album, uploader, readOnly, admin := setupMediaManageTest(t)
	ctx := auth.AddUserToContext(context.Background(), uploader)

	t.Run("happy path renames the file and updates the row", func(t *testing.T) {
		media := makeTestMediaFile(t, r, album, "20260101_120000.mp4")

		renamed, err := r.RenameMedia(ctx, media.ID, "Birthday party.mp4")
		assert.NoError(t, err)
		if !assert.NotNil(t, renamed) {
			return
		}

		wantPath := filepath.Join(album.Path, "Birthday party.mp4")
		assert.Equal(t, wantPath, renamed.Path)
		assert.Equal(t, "Birthday party.mp4", renamed.Title)
		assert.Equal(t, models.MD5Hash(wantPath), renamed.PathHash)

		_, statErr := os.Stat(wantPath)
		assert.NoError(t, statErr)
		_, statErr = os.Stat(media.Path)
		assert.True(t, os.IsNotExist(statErr), "old file should no longer exist")
	})

	t.Run("sidecar file is renamed alongside and its hash updated", func(t *testing.T) {
		media := makeTestMediaFile(t, r, album, "IMG_0001.CR2")
		sidecarPath := filepath.Join(album.Path, "IMG_0001.xmp")
		assert.NoError(t, os.WriteFile(sidecarPath, []byte("sidecar"), 0o644))
		media.SideCarPath = &sidecarPath
		assert.NoError(t, r.database.Save(media).Error)

		renamed, err := r.RenameMedia(ctx, media.ID, "Vacation.CR2")
		assert.NoError(t, err)
		if !assert.NotNil(t, renamed) || !assert.NotNil(t, renamed.SideCarPath) {
			return
		}

		wantSidecarPath := filepath.Join(album.Path, "Vacation.xmp")
		assert.Equal(t, wantSidecarPath, *renamed.SideCarPath)
		assert.NotNil(t, renamed.SideCarHash)
		assert.Equal(t, models.MD5Hash(wantSidecarPath), *renamed.SideCarHash)

		_, statErr := os.Stat(wantSidecarPath)
		assert.NoError(t, statErr)
	})

	t.Run("name collision is rejected", func(t *testing.T) {
		makeTestMediaFile(t, r, album, "existing.jpg")
		media := makeTestMediaFile(t, r, album, "to_rename.jpg")

		_, err := r.RenameMedia(ctx, media.ID, "existing.jpg")
		assert.Error(t, err)

		_, statErr := os.Stat(media.Path)
		assert.NoError(t, statErr, "original file should be untouched after a rejected rename")
	})

	t.Run("path traversal in the new name is rejected", func(t *testing.T) {
		media := makeTestMediaFile(t, r, album, "traversal_test.jpg")

		_, err := r.RenameMedia(ctx, media.ID, "../evil.jpg")
		assert.Error(t, err)
		_, statErr := os.Stat(filepath.Join(filepath.Dir(album.Path), "evil.jpg"))
		assert.True(t, os.IsNotExist(statErr))
	})

	t.Run("sidecar collision is rejected without touching the unrelated sidecar", func(t *testing.T) {
		// photo.jpg's own sidecar - must survive untouched.
		makeTestMediaFile(t, r, album, "photo.jpg")
		photoSidecarPath := filepath.Join(album.Path, "photo.xmp")
		assert.NoError(t, os.WriteFile(photoSidecarPath, []byte("photo's own sidecar"), 0o644))

		// The media file being renamed to "photo.arw" - that filename is
		// free, so the plain destination-collision check on the media file
		// itself passes, but its sidecar would land on "photo.xmp".
		media := makeTestMediaFile(t, r, album, "raw_0001.arw")
		mediaSidecarPath := filepath.Join(album.Path, "raw_0001.xmp")
		assert.NoError(t, os.WriteFile(mediaSidecarPath, []byte("raw_0001's own sidecar"), 0o644))
		media.SideCarPath = &mediaSidecarPath
		assert.NoError(t, r.database.Save(media).Error)

		_, err := r.RenameMedia(ctx, media.ID, "photo.arw")
		assert.Error(t, err)

		content, readErr := os.ReadFile(photoSidecarPath)
		assert.NoError(t, readErr)
		assert.Equal(t, "photo's own sidecar", string(content), "the unrelated sidecar must not be overwritten")

		_, statErr := os.Stat(media.Path)
		assert.NoError(t, statErr, "the renamed-away media file should have been rolled back")
		_, statErr = os.Stat(mediaSidecarPath)
		assert.NoError(t, statErr, "the media's own sidecar should still be at its original path")
	})

	t.Run("permission denied for a read-only user", func(t *testing.T) {
		media := makeTestMediaFile(t, r, album, "readonly_test.jpg")
		readOnlyCtx := auth.AddUserToContext(context.Background(), readOnly)

		_, err := r.RenameMedia(readOnlyCtx, media.ID, "should_not_work.jpg")
		assert.Error(t, err)
		_, statErr := os.Stat(media.Path)
		assert.NoError(t, statErr)
	})

	t.Run("admin bypasses the permission check", func(t *testing.T) {
		media := makeTestMediaFile(t, r, album, "admin_test.jpg")
		adminCtx := auth.AddUserToContext(context.Background(), admin)

		renamed, err := r.RenameMedia(adminCtx, media.ID, "admin_renamed.jpg")
		assert.NoError(t, err)
		assert.NotNil(t, renamed)
	})
}

func TestDeleteMedia(t *testing.T) {
	r, album, uploader, readOnly, _ := setupMediaManageTest(t)
	ctx := auth.AddUserToContext(context.Background(), uploader)

	t.Run("happy path moves the file to trash and removes the row and its dependents", func(t *testing.T) {
		media := makeTestMediaFile(t, r, album, "to_delete.jpg")
		originalPath := media.Path

		exif := models.MediaEXIF{}
		assert.NoError(t, r.database.Save(&exif).Error)
		media.ExifID = &exif.ID
		assert.NoError(t, r.database.Save(media).Error)

		ok, err := r.DeleteMedia(ctx, media.ID)
		assert.NoError(t, err)
		assert.True(t, ok)

		_, statErr := os.Stat(originalPath)
		assert.True(t, os.IsNotExist(statErr), "original file should no longer exist")

		trashEntries, err := os.ReadDir(filepath.Join(album.Path, ".trash"))
		assert.NoError(t, err)
		assert.Len(t, trashEntries, 1)

		assert.Error(t, r.database.First(&models.Media{}, media.ID).Error)
		assert.Error(t, r.database.First(&models.MediaEXIF{}, exif.ID).Error, "dependent EXIF row should cascade-delete")
	})

	t.Run("permission denied for a read-only user", func(t *testing.T) {
		media := makeTestMediaFile(t, r, album, "protected.jpg")
		readOnlyCtx := auth.AddUserToContext(context.Background(), readOnly)

		ok, err := r.DeleteMedia(readOnlyCtx, media.ID)
		assert.Error(t, err)
		assert.False(t, ok)

		_, statErr := os.Stat(media.Path)
		assert.NoError(t, statErr, "file should be untouched after a denied delete")
	})
}

func TestDeleteMediaList(t *testing.T) {
	r, album, uploader, readOnly, _ := setupMediaManageTest(t)
	ctx := auth.AddUserToContext(context.Background(), uploader)

	t.Run("mixed batch: owned files succeed independently of a denied one", func(t *testing.T) {
		media1 := makeTestMediaFile(t, r, album, "batch1.jpg")
		media2 := makeTestMediaFile(t, r, album, "batch2.jpg")

		otherAlbum := &models.Album{Title: "other_album", Path: t.TempDir()}
		assert.NoError(t, r.database.Save(otherAlbum).Error)
		assert.NoError(t, r.database.Create(&models.UserAlbums{
			UserID: readOnly.ID, AlbumID: otherAlbum.ID, Level: models.AlbumPermissionLevelRead,
		}).Error)
		deniedMedia := makeTestMediaFile(t, r, otherAlbum, "denied.jpg")

		// Run as uploader against a mix of two files they own and one in an
		// album they have no access to at all, to exercise per-item
		// success/failure in a single call.
		results, err := r.DeleteMediaList(ctx, []int{media1.ID, media2.ID, deniedMedia.ID})
		assert.NoError(t, err)
		if !assert.Len(t, results, 3) {
			return
		}

		assert.True(t, results[0].Success)
		assert.Nil(t, results[0].Error)
		assert.True(t, results[1].Success)
		assert.Nil(t, results[1].Error)
		assert.False(t, results[2].Success, "uploader has no access to otherAlbum, this item should fail")
		assert.NotNil(t, results[2].Error)

		_, statErr := os.Stat(media1.Path)
		assert.True(t, os.IsNotExist(statErr))
		_, statErr = os.Stat(media2.Path)
		assert.True(t, os.IsNotExist(statErr))
		_, statErr = os.Stat(deniedMedia.Path)
		assert.NoError(t, statErr, "the denied file must be untouched")
	})

	t.Run("empty list returns an empty result without error", func(t *testing.T) {
		results, err := r.DeleteMediaList(ctx, []int{})
		assert.NoError(t, err)
		assert.Empty(t, results)
	})
}
