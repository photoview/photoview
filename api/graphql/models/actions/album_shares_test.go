package actions_test

import (
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/graphql/models/actions"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListAlbumSharesDoesNotDiscloseSiblingTokens(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	password := "password"
	owner, err := models.RegisterUser(db, "album-shares-owner", &password, false)
	require.NoError(t, err)

	album := models.Album{Title: "shared-album", Path: "/photos/shared-album"}
	require.NoError(t, db.Create(&album).Error)
	require.NoError(t, db.Model(&album).Association("Owners").Append(owner))

	tokenA := models.ShareToken{Value: "ALBUM_TOKEN_A", OwnerID: owner.ID, AlbumID: &album.ID}
	tokenB := models.ShareToken{Value: "ALBUM_TOKEN_B", OwnerID: owner.ID, AlbumID: &album.ID}
	require.NoError(t, db.Create(&tokenA).Error)
	require.NoError(t, db.Create(&tokenB).Error)

	values := func(tokens []*models.ShareToken) []string {
		out := make([]string, len(tokens))
		for i, tok := range tokens {
			out[i] = tok.Value
		}
		return out
	}

	t.Run("share-token caller cannot read sibling tokens", func(t *testing.T) {
		got, err := actions.ListAlbumShares(db, nil, album.ID)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("unrelated user cannot read sibling tokens", func(t *testing.T) {
		other, err := models.RegisterUser(db, "album-shares-other", &password, false)
		require.NoError(t, err)
		got, err := actions.ListAlbumShares(db, other, album.ID)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("owner retains token management", func(t *testing.T) {
		got, err := actions.ListAlbumShares(db, owner, album.ID)
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"ALBUM_TOKEN_A", "ALBUM_TOKEN_B"}, values(got))
	})

	t.Run("administrator retains token management", func(t *testing.T) {
		admin, err := models.RegisterUser(db, "album-shares-admin", &password, true)
		require.NoError(t, err)
		require.True(t, admin.Admin)
		got, err := actions.ListAlbumShares(db, admin, album.ID)
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"ALBUM_TOKEN_A", "ALBUM_TOKEN_B"}, values(got))
	})
}
