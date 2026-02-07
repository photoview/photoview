package actions_test

import (
	"testing"
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/graphql/models/actions"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
)

func TestShareToken(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	password := "1234"
	user, err := models.RegisterUser(db, "user", &password, false)
	assert.NoError(t, err)

	rootAlbum := models.Album{
		Title: "root",
		Path:  "/photos",
	}

	assert.NoError(t, db.Save(&rootAlbum).Error)

	childAlbum := models.Album{
		Title:         "subalbum",
		Path:          "/photos/subalbum",
		ParentAlbumID: &rootAlbum.ID,
	}

	assert.NoError(t, db.Save(&childAlbum).Error)

	assert.NoError(t, db.Model(&user).Association("Albums").Append(&rootAlbum))
	assert.NoError(t, db.Model(&user).Association("Albums").Append(&childAlbum))

	media := []models.Media{
		{
			Title:   "pic1",
			Path:    "/photos/pic1",
			AlbumID: rootAlbum.ID,
		},
		{
			Title:   "pic3",
			Path:    "/photos/subalbum/pic3",
			AlbumID: childAlbum.ID,
		},
		{
			Title:   "pic4",
			Path:    "/photos/subalbum/pic4",
			AlbumID: childAlbum.ID,
		},
	}

	assert.NoError(t, db.Save(&media).Error)

	expireTime := time.Unix(1632866400, 0)
	sharePassword := "secretSharePassword"

	var mediaShare *models.ShareToken
	var albumShare *models.ShareToken

	t.Run("Add album share", func(t *testing.T) {
		share, err := actions.AddAlbumShare(db, user, rootAlbum.ID, &expireTime, nil)
		albumShare = share

		assert.NoError(t, err)
		assert.NotNil(t, share)

		assert.NotEmpty(t, share.Value)
		assert.Equal(t, rootAlbum.ID, *share.AlbumID)
		assert.Nil(t, share.MediaID)
	})

	t.Run("Add media share", func(t *testing.T) {
		share, err := actions.AddMediaShare(db, user, media[0].ID, &expireTime, &sharePassword)
		mediaShare = share

		assert.NoError(t, err)
		assert.NotNil(t, share)

		assert.NotEmpty(t, share.Value)
		assert.Equal(t, media[0].ID, *share.MediaID)
		assert.Nil(t, share.AlbumID)
	})

	t.Run("Delete share token", func(t *testing.T) {
		deletedShare, err := actions.DeleteShareToken(db, user.ID, mediaShare.Value)

		assert.NoError(t, err)
		assert.Equal(t, mediaShare.ID, deletedShare.ID)
	})

	t.Run("Protect share token", func(t *testing.T) {

		assert.Empty(t, albumShare.Password)

		share, err := actions.ProtectShareToken(db, user.ID, albumShare.Value, &sharePassword)
		assert.NoError(t, err)
		assert.NotEmpty(t, share.Password)

		share, err = actions.ProtectShareToken(db, user.ID, albumShare.Value, nil)
		assert.NoError(t, err)
		assert.Empty(t, share.Password)
	})

	t.Run("Set Expiration date for share token", func(t *testing.T) {
		assert.NotEmpty(t, albumShare.Expire)
		time_ := time.Date(2025, 12, 6, 0, 0, 0, 0, time.UTC)

		share, err := actions.SetExpireShareToken(db, user.ID, albumShare.Value, &time_)
		assert.NoError(t, err)
		assert.Equal(t, time_, *share.Expire)

		share, err = actions.SetExpireShareToken(db, user.ID, albumShare.Value, nil)
		assert.NoError(t, err)
		assert.Nil(t, share.Expire)
	})
}
