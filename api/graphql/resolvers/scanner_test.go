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

func TestScannerQueueStatus(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	uploader, err := models.RegisterUser(db, "queue_uploader", nil, false)
	assert.NoError(t, err)
	stranger, err := models.RegisterUser(db, "queue_stranger", nil, false)
	assert.NoError(t, err)
	admin, err := models.RegisterUser(db, "queue_admin", nil, true)
	assert.NoError(t, err)

	visibleAlbum := models.Album{Title: "visible", Path: "/photos/queue_visible"}
	assert.NoError(t, db.Save(&visibleAlbum).Error)
	assert.NoError(t, db.Create(&models.UserAlbums{
		UserID: uploader.ID, AlbumID: visibleAlbum.ID, Level: models.AlbumPermissionLevelRead,
	}).Error)

	hiddenAlbum := models.Album{Title: "hidden", Path: "/photos/queue_hidden"}
	assert.NoError(t, db.Save(&hiddenAlbum).Error)

	origGetQueueStatus := getScannerQueueStatus
	getScannerQueueStatus = func() []models.ScannerQueueItem {
		return []models.ScannerQueueItem{
			{Album: &visibleAlbum, Status: models.ScannerJobStatusRunning},
			{Album: &hiddenAlbum, Status: models.ScannerJobStatusQueued},
		}
	}
	t.Cleanup(func() { getScannerQueueStatus = origGetQueueStatus })

	r := &queryResolver{Resolver: &Resolver{database: db}}

	t.Run("a user with read access only sees the job for that album", func(t *testing.T) {
		ctx := auth.AddUserToContext(context.Background(), uploader)
		items, err := r.ScannerQueueStatus(ctx)
		assert.NoError(t, err)
		if assert.Len(t, items, 1) {
			assert.Equal(t, "visible", items[0].Album.Title)
		}
	})

	t.Run("a user with no access to either album sees nothing", func(t *testing.T) {
		ctx := auth.AddUserToContext(context.Background(), stranger)
		items, err := r.ScannerQueueStatus(ctx)
		assert.NoError(t, err)
		assert.Empty(t, items)
	})

	t.Run("admin sees every job regardless of access", func(t *testing.T) {
		ctx := auth.AddUserToContext(context.Background(), admin)
		items, err := r.ScannerQueueStatus(ctx)
		assert.NoError(t, err)
		assert.Len(t, items, 2)
	})

	t.Run("unauthenticated request is rejected", func(t *testing.T) {
		_, err := r.ScannerQueueStatus(context.Background())
		assert.Error(t, err)
	})
}
