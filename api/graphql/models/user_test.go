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

func TestUserOwnsAlbum(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	user := models.User{
		Username: "user",
	}

	if !assert.NoError(t, db.Save(&user).Error) {
		return
	}

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

	subAlbums := []models.Album{
		{
			Title:         "subalbum1",
			Path:          "/photos/album2/subalbum1",
			ParentAlbumID: &albums[1].ID,
		},
		{
			Title:         "another_sub",
			Path:          "/photos/album2/another_sub",
			ParentAlbumID: &albums[1].ID,
		},
		{
			Title:         "subalbum2",
			Path:          "/photos/album1/subalbum2",
			ParentAlbumID: &albums[0].ID,
		},
	}

	if !assert.NoError(t, db.Model(&user).Association("Albums").Append(&subAlbums)) {
		return
	}

	for _, album := range albums {
		owns, err := user.OwnsAlbum(db, &album)
		assert.NoError(t, err)
		assert.True(t, owns)
	}

	for _, album := range subAlbums {
		owns, err := user.OwnsAlbum(db, &album)
		assert.NoError(t, err)
		assert.True(t, owns)
	}

	separateAlbum := models.Album{
		Title: "separate_album",
		Path:  "/my_media/album123",
	}

	if !assert.NoError(t, db.Save(&separateAlbum).Error) {
		return
	}

	owns, err := user.OwnsAlbum(db, &separateAlbum)
	assert.NoError(t, err)
	assert.False(t, owns)
}

func TestUserCanUploadToAlbum(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	owner := models.User{Username: "owner", CanUpload: true}
	assert.NoError(t, db.Save(&owner).Error)

	nonUploader := models.User{Username: "non_uploader", CanUpload: false}
	assert.NoError(t, db.Save(&nonUploader).Error)

	admin := models.User{Username: "admin_user", Admin: true, CanUpload: false}
	assert.NoError(t, db.Save(&admin).Error)

	album := models.Album{Title: "album", Path: "/photos/album"}
	assert.NoError(t, db.Model(&owner).Association("Albums").Append(&album))

	t.Run("owner with CanUpload can upload", func(t *testing.T) {
		can, err := owner.CanUploadToAlbum(db, &album)
		assert.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("user without CanUpload cannot upload, even if they own the album", func(t *testing.T) {
		nonUploaderOwned := models.Album{Title: "album2", Path: "/photos/album2"}
		assert.NoError(t, db.Model(&nonUploader).Association("Albums").Append(&nonUploaderOwned))

		can, err := nonUploader.CanUploadToAlbum(db, &nonUploaderOwned)
		assert.NoError(t, err)
		assert.False(t, can)
	})

	t.Run("CanUpload user who does not own the album cannot upload", func(t *testing.T) {
		otherAlbum := models.Album{Title: "not_owned", Path: "/photos/not_owned"}
		assert.NoError(t, db.Save(&otherAlbum).Error)

		can, err := owner.CanUploadToAlbum(db, &otherAlbum)
		assert.NoError(t, err)
		assert.False(t, can)
	})

	t.Run("admin can upload anywhere, regardless of CanUpload/ownership", func(t *testing.T) {
		can, err := admin.CanUploadToAlbum(db, &album)
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
