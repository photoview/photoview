package models_test

import (
	"testing"
	"time"

	"github.com/kkovaletp/photoview/api/dataloader"
	"github.com/kkovaletp/photoview/api/graphql/models"
	"github.com/kkovaletp/photoview/api/test_utils"
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

	sub_albums := []models.Album{
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

	if !assert.NoError(t, db.Model(&user).Association("Albums").Append(&sub_albums)) {
		return
	}

	for _, album := range albums {
		owns, err := user.OwnsAlbum(db, &album)
		assert.NoError(t, err)
		assert.True(t, owns)
	}

	for _, album := range sub_albums {
		owns, err := user.OwnsAlbum(db, &album)
		assert.NoError(t, err)
		assert.True(t, owns)
	}

	separate_album := models.Album{
		Title: "separate_album",
		Path:  "/my_media/album123",
	}

	if !assert.NoError(t, db.Save(&separate_album).Error) {
		return
	}

	owns, err := user.OwnsAlbum(db, &separate_album)
	assert.NoError(t, err)
	assert.False(t, owns)
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
