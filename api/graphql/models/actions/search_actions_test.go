package actions_test

import (
	"fmt"
	"testing"

	"github.com/kkovaletp/photoview/api/graphql/models"
	"github.com/kkovaletp/photoview/api/graphql/models/actions"
	"github.com/kkovaletp/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
)

func TestSearch(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	user, err := models.RegisterUser(db, "user", nil, true)
	assert.NoError(t, err)

	rootAlbum := models.Album{
		Title: "root_album",
		Path:  "/media/",
	}

	assert.NoError(t, db.Create(&rootAlbum).Error)
	assert.NoError(t, db.Model(&rootAlbum).Association("Owners").Append(user))

	type Result struct {
		ID      int
		UserID  int
		AlbumID int
	}

	mediaTitles := []string{
		"SOME_IMAGE.jpg",
		"imageA.jpg",
		"imageB.jpg",
		"imageC.jpg",
		"movie.mp4",
		"person.png",
		"123.png",
		"ABC.gif",
		"dog.mov",
		"cat.mov",
		"IMG_3255.JPG",
		"IMG_5532.JPG",
		"IMG_5533.JPG",
		"IMG_5534.JPG",
		"IMG_5535.JPG",
		"IMG_5536.JPG",
	}

	for _, mediaTitle := range mediaTitles {
		image := models.Media{
			Title:   mediaTitle,
			Path:    fmt.Sprintf("/media/%s", mediaTitle),
			AlbumID: rootAlbum.ID,
		}
		assert.NoError(t, db.Create(&image).Error)
	}

	type SearchTest = struct {
		query      string
		userID     int
		limitMedia *int
		limitAlbum *int

		expectedMediaCount int
		expectedAlbumCount int
	}

	searchTests := []SearchTest{
		{
			query:              "image",
			userID:             user.ID,
			expectedMediaCount: 4,
			expectedAlbumCount: 0,
		},
		{
			query:              "g",
			userID:             user.ID,
			expectedMediaCount: 10,
			expectedAlbumCount: 0,
		},
		{
			query:              "media",
			userID:             user.ID,
			expectedMediaCount: 10,
			expectedAlbumCount: 1,
		},
	}

	for _, test := range searchTests {
		t.Run(fmt.Sprintf("Search query: '%s'", test.query), func(t *testing.T) {
			result, err := actions.Search(db, test.query, test.userID, test.limitMedia, test.limitAlbum)
			assert.NoError(t, err)

			assert.Equal(t, result.Query, test.query)
			assert.Len(t, result.Albums, test.expectedAlbumCount)
			assert.Len(t, result.Media, test.expectedMediaCount)
		})
	}
}
