package actions_test

import (
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/graphql/models/actions"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
)

func TestAlbumCover(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	rootAlbum := models.Album{
		Title: "root",
		Path:  "/photos",
	}

	if !assert.NoError(t, db.Save(&rootAlbum).Error) {
		return
	}

	children := []models.Album{
		{
			Title:         "child1",
			Path:          "/photos/child1",
			ParentAlbumID: &rootAlbum.ID,
		},
		{
			Title:         "child2",
			Path:          "/photos/child2",
			ParentAlbumID: &rootAlbum.ID,
		},
	}

	if !assert.NoError(t, db.Save(&children).Error) {
		return
	}

	photos := []models.Media{
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
			Path:    "/photos/child1/pic3",
			AlbumID: children[0].ID,
		},
		{
			Title:   "pic4",
			Path:    "/photos/child1/pic4",
			AlbumID: children[0].ID,
		},
		{
			Title:   "pic5",
			Path:    "/photos/child2/pic5",
			AlbumID: children[1].ID,
		},
		{
			Title:   "pic6",
			Path:    "/photos/child2/pic6",
			AlbumID: children[1].ID,
		},
	}

	if !assert.NoError(t, db.Save(&photos).Error) {
		return
	}

	if !assert.NoError(t, db.Model(&children[0]).Update("cover_id", &photos[3].ID).Error) {
		return
	}

	photoUrls := []models.MediaURL{
		{
			MediaID: photos[0].ID,
			Media:   &photos[0],
		},
		{
			MediaID: photos[1].ID,
			Media:   &photos[1],
		},
		{
			MediaID: photos[2].ID,
			Media:   &photos[2],
		},
		{
			MediaID: photos[3].ID,
			Media:   &photos[3],
		},
		{
			MediaID: photos[4].ID,
			Media:   &photos[4],
		},
		{
			MediaID: photos[5].ID,
			Media:   &photos[5],
		},
	}

	if !assert.NoError(t, db.Save(&photoUrls).Error) {
		return
	}

	user_pass := "password"
	regularUser, err := models.RegisterUser(db, "user1", &user_pass, false)
	if !assert.NoError(t, err) {
		return
	}

	if !assert.NoError(t, db.Model(&regularUser).Association("Albums").Append(&rootAlbum)) {
		return
	}

	if !assert.NoError(t, db.Model(&regularUser).Association("Albums").Append(&children)) {
		return
	}

	// Single test since we cannot rely on the tests being performed sequentially
	t.Run("Album get and reset cover photos", func(t *testing.T) {
		{
			album, err := actions.Album(db, regularUser, rootAlbum.ID)
			assert.NoError(t, err)

			albumThumb, err := album.Thumbnail(db)
			assert.NoError(t, err)

			// Should return pic1 since no coverID has been set
			assert.EqualValues(t, "pic1", albumThumb.Title)
		}

		{
			album, err := actions.Album(db, regularUser, children[0].ID)
			assert.NoError(t, err)

			albumThumb, err := album.Thumbnail(db)
			assert.NoError(t, err)

			// coverID has already been set
			assert.EqualValues(t, "pic4", albumThumb.Title)
		}

		resetAlbum, err := actions.ResetAlbumCover(db, regularUser, children[0].ID)
		assert.NoError(t, err)

		assert.Nil(t, resetAlbum.CoverID)

		resetThumb, err := resetAlbum.Thumbnail(db)
		assert.NoError(t, err)

		assert.Equal(t, "pic3", resetThumb.Title)
	})

	t.Run("Album change cover photos", func(t *testing.T) {
		assert.Nil(t, children[1].CoverID)

		album, err := actions.SetAlbumCover(db, regularUser, photos[4].ID)
		assert.NoError(t, err)

		assert.Equal(t, children[1].ID, album.ID)
		assert.NotNil(t, album.CoverID)
		assert.Equal(t, photos[4].ID, *album.CoverID)

		albumThumb, err := album.Thumbnail(db)
		assert.NoError(t, err)

		assert.Equal(t, photos[4].ID, albumThumb.ID)
	})

}
