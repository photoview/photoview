package actions_test

import (
	"testing"
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/graphql/models/actions"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
)

func TestMyTimeline(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	password := "1234"
	ids := make([]int, 0)
	db.Model(&models.Role{}).Where("name = ?", "USER").Pluck("id", &ids)
	user, err := models.RegisterUser(db, "user", &password, ids[0])
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
			Title:    "pic1",
			Path:     "/photos/pic1",
			AlbumID:  rootAlbum.ID,
			DateShot: time.Unix(1632758400, 0), // Sep 27 2021 16:00:00
		},
		{
			Title:    "pic2",
			Path:     "/photos/pic2",
			AlbumID:  rootAlbum.ID,
			DateShot: time.Unix(1628762400, 0), // Aug 12 2021 10:00:00
		},
		{
			Title:    "pic3",
			Path:     "/photos/subalbum/pic3",
			AlbumID:  childAlbum.ID,
			DateShot: time.Unix(1632763800, 0), // Sep 27 2021 17:30:00
		},
		{
			Title:    "pic4",
			Path:     "/photos/subalbum/pic4",
			AlbumID:  childAlbum.ID,
			DateShot: time.Unix(1628775900, 0), // Aug 12 2021 13:45:00
		},
	}

	assert.NoError(t, db.Save(&media).Error)

	_, err = user.FavoriteMedia(db, media[0].ID, true)
	assert.NoError(t, err)

	// Add media not owned by first user
	anotherUser, err := models.RegisterUser(db, "user2", &password, ids[0])
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

	t.Run("MyTimeline with no filters", func(t *testing.T) {
		timelineMedia, err := actions.MyTimeline(db, user, nil, nil, nil)

		assert.NoError(t, err)
		assert.Len(t, timelineMedia, 4)

		for i, title := range []string{"pic1", "pic3", "pic2", "pic4"} {
			assert.Equalf(t, timelineMedia[i].Title, title, "Element %d didn't match: got %s expected %s", i, timelineMedia[i].Title, title)
		}

	})

	t.Run("MyTimeline with only favorites", func(t *testing.T) {
		favorites := true
		timelineMedia, err := actions.MyTimeline(db, user, nil, &favorites, nil)

		assert.NoError(t, err)
		assert.Len(t, timelineMedia, 1)
	})

	t.Run("MyTimeline before date", func(t *testing.T) {
		beforeDate := time.Unix(1629792000, 0) // Aug 24 2021 08:00:00
		timelineMedia, err := actions.MyTimeline(db, user, nil, nil, &beforeDate)

		assert.NoError(t, err)
		assert.Len(t, timelineMedia, 2)
	})
}
