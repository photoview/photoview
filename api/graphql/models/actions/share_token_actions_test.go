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
	shareLabel := " Family album "

	var mediaShare *models.ShareToken
	var albumShare *models.ShareToken

	t.Run("Add album share", func(t *testing.T) {
		share, err := actions.AddAlbumShare(db, user, rootAlbum.ID, &expireTime, nil, &shareLabel)
		albumShare = share

		assert.NoError(t, err)
		assert.NotNil(t, share)

		assert.NotEmpty(t, share.Value)
		assert.Equal(t, "Family album", *share.Label)
		assert.Equal(t, rootAlbum.ID, *share.AlbumID)
		assert.Nil(t, share.MediaID)
	})

	t.Run("Add media share", func(t *testing.T) {
		share, err := actions.AddMediaShare(db, user, media[0].ID, &expireTime, &sharePassword, nil)
		mediaShare = share

		assert.NoError(t, err)
		assert.NotNil(t, share)

		assert.NotEmpty(t, share.Value)
		assert.Equal(t, media[0].ID, *share.MediaID)
		assert.Nil(t, share.AlbumID)
	})

	t.Run("Delete share token", func(t *testing.T) {
		deletedShare, err := actions.DeleteShareToken(db, user, mediaShare.Value)

		assert.NoError(t, err)
		assert.Equal(t, mediaShare.ID, deletedShare.ID)
	})

	t.Run("Protect share token", func(t *testing.T) {

		assert.Empty(t, albumShare.Password)

		share, err := actions.ProtectShareToken(db, user, albumShare.Value, &sharePassword)
		assert.NoError(t, err)
		assert.NotEmpty(t, share.Password)

		share, err = actions.ProtectShareToken(db, user, albumShare.Value, nil)
		assert.NoError(t, err)
		assert.Empty(t, share.Password)
	})

	t.Run("Set Expiration date for share token", func(t *testing.T) {
		assert.NotEmpty(t, albumShare.Expire)
		time_ := time.Date(2025, 12, 6, 0, 0, 0, 0, time.UTC)

		share, err := actions.SetExpireShareToken(db, user, albumShare.Value, &time_)
		assert.NoError(t, err)
		assert.Equal(t, time_, *share.Expire)

		share, err = actions.SetExpireShareToken(db, user, albumShare.Value, nil)
		assert.NoError(t, err)
		assert.Nil(t, share.Expire)
	})

	t.Run("Set share token label", func(t *testing.T) {
		label := "  Press gallery  "
		share, err := actions.SetShareTokenLabel(db, user, albumShare.Value, &label)
		assert.NoError(t, err)
		assert.Equal(t, "Press gallery", *share.Label)

		blankLabel := " "
		share, err = actions.SetShareTokenLabel(db, user, albumShare.Value, &blankLabel)
		assert.NoError(t, err)
		assert.Nil(t, share.Label)
	})
}

func TestShareTokenPermissions(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	owner, err := models.RegisterUser(db, "share_owner", nil, false)
	assert.NoError(t, err)
	reader, err := models.RegisterUser(db, "share_reader", nil, false)
	assert.NoError(t, err)
	stranger, err := models.RegisterUser(db, "share_stranger", nil, false)
	assert.NoError(t, err)
	admin, err := models.RegisterUser(db, "share_admin", nil, true)
	assert.NoError(t, err)

	album := models.Album{Title: "owners album", Path: "/photos/owner"}
	assert.NoError(t, db.Save(&album).Error)

	media := models.Media{Title: "pic", Path: "/photos/owner/pic.jpg", AlbumID: album.ID}
	assert.NoError(t, db.Save(&media).Error)

	// owner's own grant (not granted by anyone else) makes them the owner.
	assert.NoError(t, db.Create(&models.UserAlbums{
		UserID: owner.ID, AlbumID: album.ID, Level: models.AlbumPermissionLevelDelete,
	}).Error)
	// reader was granted read-only access by the owner - not an owner themselves.
	assert.NoError(t, db.Create(&models.UserAlbums{
		UserID: reader.ID, AlbumID: album.ID, Level: models.AlbumPermissionLevelRead, GrantedByUserID: &owner.ID,
	}).Error)

	// stranger has access to a completely unrelated album, but none at all
	// to the album above - this is the regression case for the bug where
	// AddAlbumShare's check wasn't scoped to the target album at all.
	otherAlbum := models.Album{Title: "strangers own album", Path: "/photos/stranger"}
	assert.NoError(t, db.Save(&otherAlbum).Error)
	assert.NoError(t, db.Create(&models.UserAlbums{
		UserID: stranger.ID, AlbumID: otherAlbum.ID, Level: models.AlbumPermissionLevelDelete,
	}).Error)

	t.Run("owner can share the album and the media", func(t *testing.T) {
		share, err := actions.AddAlbumShare(db, owner, album.ID, nil, nil, nil)
		assert.NoError(t, err)
		assert.NotNil(t, share)

		mediaShare, err := actions.AddMediaShare(db, owner, media.ID, nil, nil, nil)
		assert.NoError(t, err)
		assert.NotNil(t, mediaShare)
	})

	t.Run("a read-only recipient cannot share the album or the media", func(t *testing.T) {
		_, err := actions.AddAlbumShare(db, reader, album.ID, nil, nil, nil)
		assert.Error(t, err)

		_, err = actions.AddMediaShare(db, reader, media.ID, nil, nil, nil)
		assert.Error(t, err)
	})

	t.Run("a user with no access at all to the album cannot share it or its media", func(t *testing.T) {
		_, err := actions.AddAlbumShare(db, stranger, album.ID, nil, nil, nil)
		assert.Error(t, err)

		_, err = actions.AddMediaShare(db, stranger, media.ID, nil, nil, nil)
		assert.Error(t, err)
	})

	t.Run("admin bypasses the ownership check", func(t *testing.T) {
		share, err := actions.AddAlbumShare(db, admin, album.ID, nil, nil, nil)
		assert.NoError(t, err)
		assert.NotNil(t, share)

		mediaShare, err := actions.AddMediaShare(db, admin, media.ID, nil, nil, nil)
		assert.NoError(t, err)
		assert.NotNil(t, mediaShare)
	})
}
