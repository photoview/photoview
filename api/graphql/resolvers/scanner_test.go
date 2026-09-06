package resolvers

import (
	"context"
	"testing"

	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
)

func TestScanAlbum(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	origAddAlbumToQueue := addAlbumToQueue
	addAlbumToQueue = func(album *models.Album) error { return nil }
	t.Cleanup(func() { addAlbumToQueue = origAddAlbumToQueue })

	uploader, err := models.RegisterUser(db, "scan_uploader", nil, false)
	assert.NoError(t, err)
	readOnlyUser, err := models.RegisterUser(db, "scan_read_only", nil, false)
	assert.NoError(t, err)
	strangerUser, err := models.RegisterUser(db, "scan_stranger", nil, false)
	assert.NoError(t, err)
	admin, err := models.RegisterUser(db, "scan_admin", nil, true)
	assert.NoError(t, err)

	album := models.Album{Title: "scan_album", Path: "/photos/scan_album"}
	assert.NoError(t, db.Save(&album).Error)
	assert.NoError(t, db.Create(&models.UserAlbums{
		UserID: uploader.ID, AlbumID: album.ID, Level: models.AlbumPermissionLevelUpload,
	}).Error)
	assert.NoError(t, db.Create(&models.UserAlbums{
		UserID: readOnlyUser.ID, AlbumID: album.ID, Level: models.AlbumPermissionLevelRead,
	}).Error)

	r := &mutationResolver{Resolver: &Resolver{database: db}}

	t.Run("user with Upload level can trigger a scan", func(t *testing.T) {
		ctx := auth.AddUserToContext(context.Background(), uploader)
		result, err := r.ScanAlbum(ctx, album.ID)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("user with only Read level cannot trigger a scan", func(t *testing.T) {
		ctx := auth.AddUserToContext(context.Background(), readOnlyUser)
		_, err := r.ScanAlbum(ctx, album.ID)
		assert.Error(t, err)
	})

	t.Run("user with no access at all cannot trigger a scan", func(t *testing.T) {
		ctx := auth.AddUserToContext(context.Background(), strangerUser)
		_, err := r.ScanAlbum(ctx, album.ID)
		assert.Error(t, err)
	})

	t.Run("admin can trigger a scan regardless of grants", func(t *testing.T) {
		ctx := auth.AddUserToContext(context.Background(), admin)
		result, err := r.ScanAlbum(ctx, album.ID)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("unauthenticated request is rejected", func(t *testing.T) {
		_, err := r.ScanAlbum(context.Background(), album.ID)
		assert.Error(t, err)
	})
}
