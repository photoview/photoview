package models_test

import (
	"testing"
	"time"

	"github.com/photoview/photoview/api/dataloader"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
)

func TestUserRegistrationAuthorization(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	t.Run("Register user", func(t *testing.T) {
		password := "1234"
		user, err := models.RegisterUser(db, "admin", &password, true)
		if !assert.NoError(t, err) {
			return
		}

		assert.NotNil(t, user)
		assert.EqualValues(t, "admin", user.Username)
		assert.NotNil(t, user.Password)
		assert.NotEqualValues(t, "1234", user.Password) // should be hashed
		assert.True(t, user.Admin)
	})

	t.Run("Authorize user", func(t *testing.T) {
		user, err := models.AuthorizeUser(db, "admin", "1234")
		if !assert.NoError(t, err) {
			return
		}

		assert.NotNil(t, user)
		assert.EqualValues(t, "admin", user.Username)
	})

	t.Run("Authorize invalid credentials", func(t *testing.T) {
		user, err := models.AuthorizeUser(db, "invalid_username", "1234")
		assert.ErrorIs(t, err, models.ErrorInvalidUserCredentials)
		assert.Nil(t, user)

		user, err = models.AuthorizeUser(db, "admin", "invalid_password")
		assert.ErrorIs(t, err, models.ErrorInvalidUserCredentials)
		assert.Nil(t, user)
	})
}

func TestAccessToken(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	pass := "<hashed_password>"
	user := models.User{
		Username: "user1",
		Password: &pass,
		Admin:    false,
	}

	if !assert.NoError(t, db.Save(&user).Error) {
		return
	}

	access_token, err := user.GenerateAccessToken(db)
	if !assert.NoError(t, err) {
		return
	}

	assert.NotNil(t, access_token)
	assert.Equal(t, user.ID, access_token.UserID)
	assert.NotEmpty(t, access_token.Value)
	assert.True(t, access_token.Expire.After(time.Now()))
}

func TestUserFillAlbums(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	user := models.User{
		Username: "user",
	}

	if !assert.NoError(t, db.Save(&user).Error) {
		return
	}

	err := user.FillAlbums(db)
	assert.NoError(t, err)
	assert.Empty(t, user.Albums)

	albums := []models.Album{
		{
			Title: "album1",
			Path:  "/photos/album1",
		},
		{
			Title: "album2",
			Path:  "/photos/album2",
		},
	}

	if !assert.NoError(t, db.Model(&user).Association("Albums").Append(&albums)) {
		return
	}

	user.Albums = make([]models.Album, 0)

	err = user.FillAlbums(db)
	assert.NoError(t, err)
	assert.Len(t, user.Albums, 2)

}

func TestUserHasAlbumLevel(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	owner := models.User{Username: "owner"}
	assert.NoError(t, db.Save(&owner).Error)

	nonUploader := models.User{Username: "non_uploader"}
	assert.NoError(t, db.Save(&nonUploader).Error)

	admin := models.User{Username: "admin_user", Admin: true}
	assert.NoError(t, db.Save(&admin).Error)

	album := models.Album{Title: "album", Path: "/photos/album"}
	assert.NoError(t, db.Save(&album).Error)
	assert.NoError(t, db.Create(&models.UserAlbums{
		UserID: owner.ID, AlbumID: album.ID, Level: models.AlbumPermissionLevelUpload,
	}).Error)

	t.Run("owner with Upload level can upload", func(t *testing.T) {
		can, err := owner.HasAlbumLevel(db, &album, models.AlbumPermissionLevelUpload)
		assert.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("Upload level does not imply Delete level", func(t *testing.T) {
		can, err := owner.HasAlbumLevel(db, &album, models.AlbumPermissionLevelDelete)
		assert.NoError(t, err)
		assert.False(t, can)
	})

	t.Run("user with Read level cannot upload, even though they own the album", func(t *testing.T) {
		nonUploaderOwned := models.Album{Title: "album2", Path: "/photos/album2"}
		assert.NoError(t, db.Save(&nonUploaderOwned).Error)
		assert.NoError(t, db.Create(&models.UserAlbums{
			UserID: nonUploader.ID, AlbumID: nonUploaderOwned.ID, Level: models.AlbumPermissionLevelRead,
		}).Error)

		can, err := nonUploader.HasAlbumLevel(db, &nonUploaderOwned, models.AlbumPermissionLevelUpload)
		assert.NoError(t, err)
		assert.False(t, can)
	})

	t.Run("user who does not own the album cannot upload", func(t *testing.T) {
		otherAlbum := models.Album{Title: "not_owned", Path: "/photos/not_owned"}
		assert.NoError(t, db.Save(&otherAlbum).Error)

		can, err := owner.HasAlbumLevel(db, &otherAlbum, models.AlbumPermissionLevelUpload)
		assert.NoError(t, err)
		assert.False(t, can)
	})

	t.Run("admin can upload anywhere, regardless of grants/ownership", func(t *testing.T) {
		can, err := admin.HasAlbumLevel(db, &album, models.AlbumPermissionLevelDelete)
		assert.NoError(t, err)
		assert.True(t, can)
	})
}

func TestUserFavoriteMedia(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	user, err := models.RegisterUser(db, "user1", nil, false)
	assert.NoError(t, err)

	rootAlbum := models.Album{
		Title: "root",
		Path:  "/photos",
	}

	assert.NoError(t, db.Save(&rootAlbum).Error)
	assert.NoError(t, db.Model(&user).Association("Albums").Append(&rootAlbum))

	media := models.Media{
		Title:   "pic1",
		Path:    "/photos/pic1",
		AlbumID: rootAlbum.ID,
	}

	assert.NoError(t, db.Save(&media).Error)

	// test that it starts out being false
	favourite, err := dataloader.NewUserFavoriteLoader(db).Load(&models.UserMediaData{
		UserID:  user.ID,
		MediaID: media.ID,
	})

	assert.NoError(t, err)
	assert.False(t, favourite)

	favMedia, err := user.FavoriteMedia(db, media.ID, true)
	assert.NoError(t, err)
	assert.NotNil(t, favMedia)

	// test that it is now true
	favourite, err = dataloader.NewUserFavoriteLoader(db).Load(&models.UserMediaData{
		UserID:  user.ID,
		MediaID: media.ID,
	})

	assert.NoError(t, err)
	assert.True(t, favourite)
}

func TestUserHideAlbum(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	userA, err := models.RegisterUser(db, "hide_user_a", nil, false)
	assert.NoError(t, err)
	userB, err := models.RegisterUser(db, "hide_user_b", nil, false)
	assert.NoError(t, err)

	album := models.Album{Title: "album", Path: "/photos/hide_test"}
	assert.NoError(t, db.Save(&album).Error)

	loadHidden := func(user *models.User) bool {
		hidden, err := dataloader.NewAlbumHiddenLoader(db).Load(&models.UserAlbumData{
			UserID:  user.ID,
			AlbumID: album.ID,
		})
		assert.NoError(t, err)
		return hidden
	}

	assert.False(t, loadHidden(userA))

	_, err = userA.HideAlbum(db, album.ID, true)
	assert.NoError(t, err)
	assert.True(t, loadHidden(userA))
	assert.False(t, loadHidden(userB), "hiding for one user must not affect another")

	_, err = userA.HideAlbum(db, album.ID, false)
	assert.NoError(t, err)
	assert.False(t, loadHidden(userA))
}

func TestUnhideAllAlbums(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	user, err := models.RegisterUser(db, "unhide_all_user", nil, false)
	assert.NoError(t, err)

	album1 := models.Album{Title: "album1", Path: "/photos/unhide1"}
	assert.NoError(t, db.Save(&album1).Error)
	album2 := models.Album{Title: "album2", Path: "/photos/unhide2"}
	assert.NoError(t, db.Save(&album2).Error)

	_, err = user.HideAlbum(db, album1.ID, true)
	assert.NoError(t, err)
	_, err = user.HideAlbum(db, album2.ID, true)
	assert.NoError(t, err)

	assert.NoError(t, user.UnhideAllAlbums(db))

	var hiddenCount int64
	assert.NoError(t, db.Model(&models.UserAlbumData{}).
		Where("user_id = ? AND hidden = true", user.ID).
		Count(&hiddenCount).Error)
	assert.Zero(t, hiddenCount)
}
