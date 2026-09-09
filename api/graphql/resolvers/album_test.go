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

func TestAlbumTreeChildren(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	user, err := models.RegisterUser(db, "tree_children_user", nil, false)
	assert.NoError(t, err)

	rootA := models.Album{Title: "root_a", Path: "/photos/tree_root_a"}
	assert.NoError(t, db.Save(&rootA).Error)
	rootB := models.Album{Title: "root_b", Path: "/photos/tree_root_b"}
	assert.NoError(t, db.Save(&rootB).Error)

	childA1 := models.Album{Title: "child_a1", Path: "/photos/tree_root_a/child1", ParentAlbumID: &rootA.ID}
	assert.NoError(t, db.Save(&childA1).Error)
	childA2 := models.Album{Title: "child_a2", Path: "/photos/tree_root_a/child2", ParentAlbumID: &rootA.ID}
	assert.NoError(t, db.Save(&childA2).Error)
	hiddenChildA := models.Album{Title: "hidden_child_a", Path: "/photos/tree_root_a/hidden", ParentAlbumID: &rootA.ID}
	assert.NoError(t, db.Save(&hiddenChildA).Error)
	childB1 := models.Album{Title: "child_b1", Path: "/photos/tree_root_b/child1", ParentAlbumID: &rootB.ID}
	assert.NoError(t, db.Save(&childB1).Error)

	// PropagateAlbumLevel (rather than a single UserAlbums row on the
	// roots) so the already-existing children get their own grant row too,
	// matching the real invariant AlbumTreeChildren's authorization check
	// relies on: every album a user can reach has its own row for them.
	assert.NoError(t, models.PropagateAlbumLevel(db, rootA.ID, user.ID, models.AlbumPermissionLevelRead, nil))
	assert.NoError(t, models.PropagateAlbumLevel(db, rootB.ID, user.ID, models.AlbumPermissionLevelRead, nil))
	_, err = user.HideAlbum(db, hiddenChildA.ID, true)
	assert.NoError(t, err)

	r := &queryResolver{Resolver: &Resolver{database: db}}
	ctx := auth.AddUserToContext(context.Background(), user)

	t.Run("returns grouped children for multiple album ids in one call", func(t *testing.T) {
		results, err := r.AlbumTreeChildren(ctx, []int{rootA.ID, rootB.ID}, nil)
		assert.NoError(t, err)
		if assert.Len(t, results, 2) {
			byAlbum := make(map[int]*models.AlbumTreeChildren, len(results))
			for _, res := range results {
				byAlbum[res.AlbumID] = res
			}

			if assert.Contains(t, byAlbum, rootA.ID) {
				assert.Len(t, byAlbum[rootA.ID].Children, 2, "hidden child excluded by default")
			}
			if assert.Contains(t, byAlbum, rootB.ID) {
				assert.Len(t, byAlbum[rootB.ID].Children, 1)
			}
		}
	})

	t.Run("showHidden reveals the hidden child too", func(t *testing.T) {
		showHidden := true
		results, err := r.AlbumTreeChildren(ctx, []int{rootA.ID}, &showHidden)
		assert.NoError(t, err)
		if assert.Len(t, results, 1) {
			assert.Len(t, results[0].Children, 3)
		}
	})

	t.Run("an album with no children returns an empty slice, not an error", func(t *testing.T) {
		results, err := r.AlbumTreeChildren(ctx, []int{childA1.ID}, nil)
		assert.NoError(t, err)
		if assert.Len(t, results, 1) {
			assert.Empty(t, results[0].Children)
		}
	})

	t.Run("an empty id list short-circuits to an empty result", func(t *testing.T) {
		results, err := r.AlbumTreeChildren(ctx, []int{}, nil)
		assert.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("a non-admin cannot read children of an album they have no access to", func(t *testing.T) {
		stranger, err := models.RegisterUser(db, "tree_children_stranger", nil, false)
		assert.NoError(t, err)
		strangerCtx := auth.AddUserToContext(context.Background(), stranger)

		results, err := r.AlbumTreeChildren(strangerCtx, []int{rootA.ID, rootB.ID}, nil)
		assert.NoError(t, err)
		if assert.Len(t, results, 2) {
			for _, res := range results {
				assert.Empty(t, res.Children, "a user with no grant on the requested album must not see its children")
			}
		}
	})

	t.Run("admin sees children regardless of their own grants", func(t *testing.T) {
		admin, err := models.RegisterUser(db, "tree_children_admin", nil, true)
		assert.NoError(t, err)
		adminCtx := auth.AddUserToContext(context.Background(), admin)

		results, err := r.AlbumTreeChildren(adminCtx, []int{rootA.ID}, nil)
		assert.NoError(t, err)
		if assert.Len(t, results, 1) {
			// "Hidden" is a per-viewer preference (UserAlbumData.Hidden) -
			// this admin never hid hiddenChildA themselves, so unlike
			// `user` earlier, they see all three children by default.
			assert.Len(t, results[0].Children, 3)
		}
	})

	t.Run("unauthenticated request is rejected", func(t *testing.T) {
		_, err := r.AlbumTreeChildren(context.Background(), []int{rootA.ID}, nil)
		assert.Error(t, err)
	})
}
