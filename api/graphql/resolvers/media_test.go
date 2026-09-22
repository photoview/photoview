package resolvers

import (
	"context"
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/graphql/models/actions"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
)

// An album share token has no MediaID; queryResolver.Media used to dereference
// shareToken.MediaID unconditionally, which panicked when a caller supplied
// album-token credentials to the media query instead of a media token.
func TestMediaWithAlbumTokenCredentials(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	password := "1234"
	owner, err := models.RegisterUser(db, "media-token-owner", &password, false)
	assert.NoError(t, err)

	album := models.Album{
		Title: "root",
		Path:  "/photos",
	}
	assert.NoError(t, db.Save(&album).Error)
	assert.NoError(t, db.Model(&owner).Association("Albums").Append(&album))

	media := models.Media{
		Title:   "pic1",
		Path:    "/photos/pic1",
		AlbumID: album.ID,
	}
	assert.NoError(t, db.Save(&media).Error)

	albumShare, err := actions.AddAlbumShare(db, owner, album.ID, nil, nil, nil)
	assert.NoError(t, err)

	resolver := &queryResolver{Resolver: &Resolver{database: db}}

	t.Run("album token credentials do not panic and are rejected", func(t *testing.T) {
		assert.NotPanics(t, func() {
			result, err := resolver.Media(context.Background(), media.ID, &models.ShareTokenCredentials{
				Token: albumShare.Value,
			})
			assert.Error(t, err)
			assert.Nil(t, result)
		})
	})

	mediaShare, err := actions.AddMediaShare(db, owner, media.ID, nil, nil, nil)
	assert.NoError(t, err)

	t.Run("matching media token credentials still work", func(t *testing.T) {
		result, err := resolver.Media(context.Background(), media.ID, &models.ShareTokenCredentials{
			Token: mediaShare.Value,
		})
		assert.NoError(t, err)
		if assert.NotNil(t, result) {
			assert.Equal(t, media.ID, result.ID)
		}
	})
}
