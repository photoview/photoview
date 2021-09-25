package actions_test

import (
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/graphql/models/actions"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
)

func TestMyMedia(t *testing.T) {
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
			Title:   "pic2",
			Path:    "/photos/pic2",
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

	anotherUser, err := models.RegisterUser(db, "user2", &password, false)
	assert.NoError(t, err)

	anotherAlbum := models.Album{
		Title: "AnotherAlbum",
		Path:  "/another",
	}

	assert.NoError(t, db.Save(&anotherAlbum).Error)

	anotherMedia := models.Media{
		Title:   "anotherPic",
		Path:    "/another/anotherPic",
		AlbumID: anotherAlbum.ID,
	}

	assert.NoError(t, db.Save(&anotherMedia).Error)

	assert.NoError(t, db.Model(&anotherUser).Association("Albums").Append(&anotherAlbum))

	t.Run("Simple query", func(t *testing.T) {
		myMedia, err := actions.MyMedia(db, user, nil, nil)

		assert.NoError(t, err)
		assert.Len(t, myMedia, 4)
	})
}
