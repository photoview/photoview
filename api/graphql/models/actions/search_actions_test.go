package actions_test

import (
	"fmt"
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/graphql/models/actions"
	"github.com/photoview/photoview/api/test_utils"
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

	noLimit := 0

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
		{
			query:              "g",
			userID:             user.ID,
			limitMedia:         &noLimit,
			expectedMediaCount: 14,
			expectedAlbumCount: 0,
		},
	}

	for _, test := range searchTests {
		t.Run(fmt.Sprintf("Search query: '%s'", test.query), func(t *testing.T) {
			result, err := actions.Search(db, test.query, test.userID, test.limitMedia, test.limitAlbum, nil)
			assert.NoError(t, err)

			assert.Equal(t, result.Query, test.query)
			assert.Len(t, result.Albums, test.expectedAlbumCount)
			assert.Len(t, result.Media, test.expectedMediaCount)
		})
	}
}

func TestSearchAlbumOrder(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	user, err := models.RegisterUser(db, "user", nil, true)
	assert.NoError(t, err)

	albumTitles := []string{
		"gallery_2024-01",
		"gallery_2023-12",
		"gallery_2024-02",
	}

	for _, title := range albumTitles {
		album := models.Album{
			Title: title,
			Path:  fmt.Sprintf("/media/%s", title),
		}
		assert.NoError(t, db.Create(&album).Error)
		assert.NoError(t, db.Model(&album).Association("Owners").Append(user))
	}

	result, err := actions.Search(db, "gallery", user.ID, nil, nil, nil)
	assert.NoError(t, err)

	var titles []string
	for _, album := range result.Albums {
		titles = append(titles, album.Title)
	}

	// Matches are ordered alphabetically descending by title, so
	// Year/Year-Month-named albums come back newest first.
	assert.Equal(t, []string{"gallery_2024-02", "gallery_2024-01", "gallery_2023-12"}, titles)
}

func TestSearchExcludesHiddenAlbumsAndDescendants(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	user, err := models.RegisterUser(db, "search_hide_user", nil, false)
	assert.NoError(t, err)

	root := models.Album{Title: "search_root", Path: "/media/search_root"}
	assert.NoError(t, db.Create(&root).Error)
	hiddenChild := models.Album{Title: "search_hidden_child", Path: "/media/search_root/hidden_child", ParentAlbumID: &root.ID}
	assert.NoError(t, db.Create(&hiddenChild).Error)
	grandchild := models.Album{Title: "search_needle_grandchild", Path: "/media/search_root/hidden_child/grandchild", ParentAlbumID: &hiddenChild.ID}
	assert.NoError(t, db.Create(&grandchild).Error)
	unrelated := models.Album{Title: "search_needle_unrelated", Path: "/media/search_needle_unrelated"}
	assert.NoError(t, db.Create(&unrelated).Error)

	for _, album := range []models.Album{root, hiddenChild, grandchild, unrelated} {
		assert.NoError(t, db.Create(&models.UserAlbums{
			UserID: user.ID, AlbumID: album.ID, Level: models.AlbumPermissionLevelRead,
		}).Error)
	}

	_, err = user.HideAlbum(db, hiddenChild.ID, true)
	assert.NoError(t, err)

	t.Run("hidden ancestor excludes its descendant from search, even though the descendant itself isn't marked hidden", func(t *testing.T) {
		result, err := actions.Search(db, "search_needle", user.ID, nil, nil, nil)
		assert.NoError(t, err)

		var titles []string
		for _, album := range result.Albums {
			titles = append(titles, album.Title)
		}
		assert.Contains(t, titles, "search_needle_unrelated")
		assert.NotContains(t, titles, "search_needle_grandchild")
	})

	t.Run("showHidden restores it", func(t *testing.T) {
		showHidden := true
		result, err := actions.Search(db, "search_needle", user.ID, nil, nil, &showHidden)
		assert.NoError(t, err)

		var titles []string
		for _, album := range result.Albums {
			titles = append(titles, album.Title)
		}
		assert.Contains(t, titles, "search_needle_unrelated")
		assert.Contains(t, titles, "search_needle_grandchild")
	})
}
