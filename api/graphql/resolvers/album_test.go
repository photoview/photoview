package resolvers

import (
	"context"
	"fmt"
	"testing"

	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestAlbumTreeChildren(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	ownedRoot := models.Album{Title: "owned_root", Path: "/photos/owned"}
	assert.NoError(t, db.Save(&ownedRoot).Error)
	ownedChild := models.Album{Title: "b_child", Path: "/photos/owned/b", ParentAlbumID: &ownedRoot.ID}
	assert.NoError(t, db.Save(&ownedChild).Error)
	ownedChild2 := models.Album{Title: "a_child", Path: "/photos/owned/a", ParentAlbumID: &ownedRoot.ID}
	assert.NoError(t, db.Save(&ownedChild2).Error)

	foreignRoot := models.Album{Title: "foreign_root", Path: "/photos/foreign"}
	assert.NoError(t, db.Save(&foreignRoot).Error)
	foreignChild := models.Album{Title: "secret", Path: "/photos/foreign/secret", ParentAlbumID: &foreignRoot.ID}
	assert.NoError(t, db.Save(&foreignChild).Error)

	user, err := models.RegisterUser(db, "tree_user", nil, false)
	assert.NoError(t, err)
	assert.NoError(t, db.Model(&user).Association("Albums").Append(&ownedRoot, &ownedChild, &ownedChild2))

	admin, err := models.RegisterUser(db, "tree_admin", nil, true)
	assert.NoError(t, err)

	r := &queryResolver{Resolver: &Resolver{database: db}}

	t.Run("children come back grouped by parent and sorted by title", func(t *testing.T) {
		result, err := r.AlbumTreeChildren(auth.AddUserToContext(context.Background(), user), []int{ownedRoot.ID})
		assert.NoError(t, err)

		if assert.Len(t, result, 1) {
			assert.Equal(t, ownedRoot.ID, result[0].AlbumID)
			if assert.Len(t, result[0].Children, 2) {
				assert.Equal(t, "a_child", result[0].Children[0].Title)
				assert.Equal(t, "b_child", result[0].Children[1].Title)
			}
		}
	})

	t.Run("an album id the caller has no access to yields nothing", func(t *testing.T) {
		// The ids come straight from the client here, unlike the subAlbums
		// field, so this is the check that stops someone reading another
		// user's folder names by guessing ids.
		result, err := r.AlbumTreeChildren(auth.AddUserToContext(context.Background(), user), []int{foreignRoot.ID})
		assert.NoError(t, err)

		if assert.Len(t, result, 1) {
			assert.Equal(t, foreignRoot.ID, result[0].AlbumID)
			assert.Empty(t, result[0].Children)
		}
	})

	t.Run("mixing own and foreign ids answers only for the own ones", func(t *testing.T) {
		// The interesting attack is not a request for someone else's album on
		// its own - it is one hidden among ids the caller is allowed to ask
		// about, in the hope that the authorization is done once for the batch
		// rather than per id.
		result, err := r.AlbumTreeChildren(
			auth.AddUserToContext(context.Background(), user),
			[]int{ownedRoot.ID, foreignRoot.ID})
		assert.NoError(t, err)

		if assert.Len(t, result, 2) {
			assert.Equal(t, ownedRoot.ID, result[0].AlbumID)
			assert.Len(t, result[0].Children, 2)

			// Same shape as an album that simply has no children: the reply
			// says nothing about whether the album exists, so it cannot be
			// used to probe for other users' folders either.
			assert.Equal(t, foreignRoot.ID, result[1].AlbumID)
			assert.Empty(t, result[1].Children)
		}
	})

	t.Run("an id that does not exist looks the same as one without access", func(t *testing.T) {
		missingID := foreignChild.ID + 1000

		result, err := r.AlbumTreeChildren(
			auth.AddUserToContext(context.Background(), user), []int{missingID})
		assert.NoError(t, err)

		if assert.Len(t, result, 1) {
			assert.Equal(t, missingID, result[0].AlbumID)
			assert.Empty(t, result[0].Children)
		}
	})

	t.Run("an admin sees children of any album", func(t *testing.T) {
		result, err := r.AlbumTreeChildren(auth.AddUserToContext(context.Background(), admin), []int{foreignRoot.ID})
		assert.NoError(t, err)

		if assert.Len(t, result, 1) {
			assert.Len(t, result[0].Children, 1)
		}
	})

	t.Run("an empty request is answered without touching the database", func(t *testing.T) {
		result, err := r.AlbumTreeChildren(auth.AddUserToContext(context.Background(), user), []int{})
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("access inherited from an ancestor counts too", func(t *testing.T) {
		// A user linked only to a root owns everything below it, which the
		// direct user_albums lookup alone would miss.
		heir, err := models.RegisterUser(db, "tree_heir", nil, false)
		assert.NoError(t, err)
		assert.NoError(t, db.Model(&heir).Association("Albums").Append(&ownedRoot))

		result, err := r.AlbumTreeChildren(auth.AddUserToContext(context.Background(), heir), []int{ownedChild.ID})
		assert.NoError(t, err)

		if assert.Len(t, result, 1) {
			assert.Equal(t, ownedChild.ID, result[0].AlbumID)
		}
	})

	t.Run("an oversized request is refused rather than built into a huge query", func(t *testing.T) {
		tooMany := make([]int, maxAlbumTreeChildrenIDs+1)
		for i := range tooMany {
			tooMany[i] = ownedRoot.ID
		}

		_, err := r.AlbumTreeChildren(auth.AddUserToContext(context.Background(), user), tooMany)
		assert.Error(t, err)
	})

	t.Run("an unauthenticated request is refused", func(t *testing.T) {
		_, err := r.AlbumTreeChildren(context.Background(), []int{ownedRoot.ID})
		assert.Error(t, err)
	})
}

func TestAuthorizedAlbumIDsIsOneQueryForAnyBatchSize(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	// A user linked only to the root owns everything below it, but none of the
	// albums below has a user_albums row of its own - the case that used to
	// cost up to two queries per id.
	root := models.Album{Title: "root", Path: "/photos/batch"}
	assert.NoError(t, db.Save(&root).Error)

	requested := make([]int, 0, 120)

	for i := 0; i < 60; i++ {
		child := models.Album{Title: fmt.Sprintf("child_%02d", i), Path: fmt.Sprintf("/photos/batch/%02d", i), ParentAlbumID: &root.ID}
		assert.NoError(t, db.Save(&child).Error)

		grandchild := models.Album{Title: "leaf", Path: fmt.Sprintf("/photos/batch/%02d/leaf", i), ParentAlbumID: &child.ID}
		assert.NoError(t, db.Save(&grandchild).Error)

		requested = append(requested, child.ID, grandchild.ID)
	}

	user, err := models.RegisterUser(db, "batch_user", nil, false)
	assert.NoError(t, err)
	assert.NoError(t, db.Model(&user).Association("Albums").Append(&root))

	var statements int

	count := func(*gorm.DB) { statements++ }
	for _, chain := range []interface {
		Register(name string, fn func(*gorm.DB)) error
	}{db.Callback().Query().After("gorm:query"), db.Callback().Row().After("gorm:row"), db.Callback().Raw().After("gorm:raw")} {
		assert.NoError(t, chain.Register("count_statements", count))
	}

	t.Cleanup(func() {
		_ = db.Callback().Query().Remove("count_statements")
		_ = db.Callback().Row().Remove("count_statements")
		_ = db.Callback().Raw().Remove("count_statements")
	})

	r := &queryResolver{Resolver: &Resolver{database: db}}

	authorized, err := r.authorizedAlbumIDs(context.Background(), user, requested)
	assert.NoError(t, err)

	assert.Equal(t, requested, authorized, "every inherited album is authorized, in the order asked for")
	assert.Equal(t, 1, statements, "authorizing a batch must not cost a query per album")
}
