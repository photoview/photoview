package resolvers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/photoview/photoview/api/dataloader"
	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// contextWithLoaders returns ctx augmented with a real set of dataloaders,
// the way an actual request would have them attached by dataloader.Middleware,
// so resolvers using dataloader.For(ctx) can be exercised directly in tests.
func contextWithLoaders(db *gorm.DB, ctx context.Context) context.Context {
	var loadedCtx context.Context
	handler := dataloader.Middleware(db)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loadedCtx = r.Context()
	}))
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil).WithContext(ctx))
	return loadedCtx
}

func TestAlbumViewerPermissionFields(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	owner, err := models.RegisterUser(db, "album_viewer_owner", nil, false)
	assert.NoError(t, err)
	uploader, err := models.RegisterUser(db, "album_viewer_uploader", nil, false)
	assert.NoError(t, err)
	reader, err := models.RegisterUser(db, "album_viewer_reader", nil, false)
	assert.NoError(t, err)
	stranger, err := models.RegisterUser(db, "album_viewer_stranger", nil, false)
	assert.NoError(t, err)
	admin, err := models.RegisterUser(db, "album_viewer_admin", nil, true)
	assert.NoError(t, err)

	album := models.Album{Title: "album", Path: "/photos/viewer_fields"}
	assert.NoError(t, db.Save(&album).Error)

	assert.NoError(t, db.Create(&models.UserAlbums{
		UserID: owner.ID, AlbumID: album.ID, Level: models.AlbumPermissionLevelDelete,
	}).Error)
	assert.NoError(t, db.Create(&models.UserAlbums{
		UserID: uploader.ID, AlbumID: album.ID, Level: models.AlbumPermissionLevelUpload, GrantedByUserID: &owner.ID,
	}).Error)
	assert.NoError(t, db.Create(&models.UserAlbums{
		UserID: reader.ID, AlbumID: album.ID, Level: models.AlbumPermissionLevelRead, GrantedByUserID: &owner.ID,
	}).Error)

	r := &albumResolver{Resolver: &Resolver{database: db}}

	ctxFor := func(user *models.User) context.Context {
		return contextWithLoaders(db, auth.AddUserToContext(context.Background(), user))
	}

	t.Run("owner can upload, delete, and is the owner", func(t *testing.T) {
		ctx := ctxFor(owner)
		canUpload, err := r.ViewerCanUpload(ctx, &album)
		assert.NoError(t, err)
		assert.True(t, canUpload)

		canDelete, err := r.ViewerCanDelete(ctx, &album)
		assert.NoError(t, err)
		assert.True(t, canDelete)

		isOwner, err := r.ViewerIsOwner(ctx, &album)
		assert.NoError(t, err)
		assert.True(t, isOwner)
	})

	t.Run("upload-level recipient can upload but not delete, and is not the owner", func(t *testing.T) {
		ctx := ctxFor(uploader)
		canUpload, err := r.ViewerCanUpload(ctx, &album)
		assert.NoError(t, err)
		assert.True(t, canUpload)

		canDelete, err := r.ViewerCanDelete(ctx, &album)
		assert.NoError(t, err)
		assert.False(t, canDelete)

		isOwner, err := r.ViewerIsOwner(ctx, &album)
		assert.NoError(t, err)
		assert.False(t, isOwner)
	})

	t.Run("read-only recipient can neither upload nor delete", func(t *testing.T) {
		ctx := ctxFor(reader)
		canUpload, err := r.ViewerCanUpload(ctx, &album)
		assert.NoError(t, err)
		assert.False(t, canUpload)

		canDelete, err := r.ViewerCanDelete(ctx, &album)
		assert.NoError(t, err)
		assert.False(t, canDelete)
	})

	t.Run("a user with no grant at all gets false for every field", func(t *testing.T) {
		ctx := ctxFor(stranger)
		canUpload, err := r.ViewerCanUpload(ctx, &album)
		assert.NoError(t, err)
		assert.False(t, canUpload)

		canDelete, err := r.ViewerCanDelete(ctx, &album)
		assert.NoError(t, err)
		assert.False(t, canDelete)

		isOwner, err := r.ViewerIsOwner(ctx, &album)
		assert.NoError(t, err)
		assert.False(t, isOwner)
	})

	t.Run("admin bypasses grants entirely", func(t *testing.T) {
		ctx := ctxFor(admin)
		canUpload, err := r.ViewerCanUpload(ctx, &album)
		assert.NoError(t, err)
		assert.True(t, canUpload)

		canDelete, err := r.ViewerCanDelete(ctx, &album)
		assert.NoError(t, err)
		assert.True(t, canDelete)

		isOwner, err := r.ViewerIsOwner(ctx, &album)
		assert.NoError(t, err)
		assert.True(t, isOwner)
	})

	t.Run("unauthenticated viewer gets false for every field", func(t *testing.T) {
		ctx := contextWithLoaders(db, context.Background())
		canUpload, err := r.ViewerCanUpload(ctx, &album)
		assert.NoError(t, err)
		assert.False(t, canUpload)
	})
}
