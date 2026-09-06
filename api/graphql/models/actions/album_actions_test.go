package actions_test

import (
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/graphql/models/actions"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
)

func TestAlbumPath(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	album := models.Album{
		Title: "Three",
		Path:  "/one/two/three",
		ParentAlbum: &models.Album{
			Title: "Two",
			Path:  "/one/two",
			ParentAlbum: &models.Album{
				Title: "One",
				Path:  "/one",
			},
		},
	}

	assert.NoError(t, db.Save(&album).Error)

	user, err := models.RegisterUser(db, "user", nil, false)
	assert.NoError(t, err)

	// Grant access on the whole chain, matching how a real grant on "One"
	// would actually leave the database: PropagateAlbumLevel/the scanner's
	// copy-on-create logic both cascade a grant down onto every existing
	// descendant, so "Two" and "Three" get their own rows too, not just
	// "One". (The original version of this test only appended the
	// grandparent, relying on the old OwnsAlbum's ancestor-walk to infer
	// access to its descendants - a shape that never occurs in production.)
	assert.NoError(t, db.Model(&user).Association("Albums").Append(&album, album.ParentAlbum, album.ParentAlbum.ParentAlbum))

	albumPath, err := actions.AlbumPath(db, user, &album)
	assert.NoError(t, err)
	assert.Len(t, albumPath, 2)
	assert.Equal(t, "Two", albumPath[0].Title)
	assert.Equal(t, "One", albumPath[1].Title)
}

func TestAlbumForbidden(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	album := models.Album{Title: "private", Path: "/photos/private"}
	assert.NoError(t, db.Save(&album).Error)

	stranger, err := models.RegisterUser(db, "album_stranger", nil, false)
	assert.NoError(t, err)

	_, err = actions.Album(db, stranger, album.ID)
	assert.Error(t, err)
}

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

	userPass := "password"
	regularUser, err := models.RegisterUser(db, "user1", &userPass, false)
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

			// Should return the latest photo since no coverID has been set
			assert.EqualValues(t, "pic6", albumThumb.Title)
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

		assert.Equal(t, "pic4", resetThumb.Title)
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

func TestAlbumsSingleRootExpand(t *testing.T) {
	db := test_utils.DatabaseTest(t)
	boolFalse := false
	boolTrue := true

	unrelatedAlbum := models.Album{
		Title: "unrelated_album",
		Path:  "/another_place",
	}
	err := db.Create(&unrelatedAlbum).Error
	assert.NoError(t, err)

	user, err := models.RegisterUser(db, "user", nil, false)
	assert.NoError(t, err)

	rootAlbum := models.Album{
		Title: "root",
		Path:  "/root",
	}

	err = db.Model(&user).Association("Albums").Replace(&rootAlbum)
	assert.NoError(t, err)

	t.Run("Single root album, no children", func(t *testing.T) {
		returnedAlbums, err := actions.MyAlbums(db, user, nil, nil, &boolTrue, &boolTrue, &boolFalse, nil)
		assert.NoError(t, err)

		assert.Len(t, returnedAlbums, 1)
	})

	childAlbums := []models.Album{
		{
			Title:         "child1",
			Path:          "/root/child1",
			ParentAlbumID: &rootAlbum.ID,
		},
		{
			Title:         "child2",
			Path:          "/root/child2",
			ParentAlbumID: &rootAlbum.ID,
		},
		{
			Title:         "child3",
			Path:          "/root/child3",
			ParentAlbumID: &rootAlbum.ID,
		},
	}

	err = db.Model(&user).Association("Albums").Append(&childAlbums)
	assert.NoError(t, err)

	t.Run("Single root album, multiple children", func(t *testing.T) {

		returnedAlbums, err := actions.MyAlbums(db, user, nil, nil, &boolTrue, &boolTrue, &boolFalse, nil)
		assert.NoError(t, err)

		assert.Len(t, returnedAlbums, 3)
	})

}

// Related to #658
func TestNonRootAlbumPath(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	boolTrue := true
	boolFalse := false

	rootAlbum := models.Album{
		Title: "root",
		Path:  "/root",
	}

	childAlbum := models.Album{
		Title:         "child",
		Path:          "/root/child",
		ParentAlbumID: &rootAlbum.ID,
	}

	assert.NoError(t, db.Create(&rootAlbum).Error)

	// Register user
	user, err := models.RegisterUser(db, "user", nil, false)
	assert.NoError(t, err)

	// Assign album to user
	err = db.Model(&user).Association("Albums").Append(&childAlbum)
	assert.NoError(t, err)

	// The child album is a "local root album" for the user, as it does not have access to the root album
	t.Run("User should only see child album", func(t *testing.T) {
		returnedAlbums, err := actions.MyAlbums(db, user, nil, nil, &boolTrue, &boolTrue, &boolFalse, nil)
		assert.NoError(t, err)

		assert.Len(t, returnedAlbums, 1)
		assert.Equal(t, "child", returnedAlbums[0].Title)
	})
}

// Related to #658
func TestNonRootAlbumPathMultipleUsers(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	boolTrue := true
	boolFalse := false

	rootAlbum := models.Album{
		Title: "root",
		Path:  "/root",
	}

	child1Album := models.Album{
		Title:         "child1",
		Path:          "/root/child1",
		ParentAlbumID: &rootAlbum.ID,
	}

	child2Album := models.Album{
		Title:         "child2",
		Path:          "/root/child2",
		ParentAlbumID: &rootAlbum.ID,
	}

	// Admin should have access to all albums
	adminAlbums := []*models.Album{&rootAlbum, &child1Album, &child2Album}

	assert.NoError(t, db.Create(&rootAlbum).Error)

	// Register users
	admin, err := models.RegisterUser(db, "admin", nil, false)
	assert.NoError(t, err)

	user1, err := models.RegisterUser(db, "user1", nil, false)
	assert.NoError(t, err)

	user2, err := models.RegisterUser(db, "user2", nil, false)
	assert.NoError(t, err)

	// Assign albums to users
	err = db.Model(&admin).Association("Albums").Append(&adminAlbums)
	assert.NoError(t, err)

	err = db.Model(&user1).Association("Albums").Append(&child1Album)
	assert.NoError(t, err)

	err = db.Model(&user2).Association("Albums").Append(&child2Album)
	assert.NoError(t, err)

	t.Run("Admin should see all albums", func(t *testing.T) {
		returnedAlbums, err := actions.MyAlbums(db, admin, nil, nil, &boolTrue, &boolTrue, &boolFalse, nil)
		assert.NoError(t, err)

		assert.Len(t, returnedAlbums, 2)
		assert.Equal(t, "child1", returnedAlbums[0].Title)
		assert.Equal(t, "child2", returnedAlbums[1].Title)
	})

	t.Run("User 1 should only see child1 album", func(t *testing.T) {
		returnedAlbums, err := actions.MyAlbums(db, user1, nil, nil, &boolTrue, &boolTrue, &boolFalse, nil)
		assert.NoError(t, err)

		assert.Len(t, returnedAlbums, 1)
		assert.Equal(t, "child1", returnedAlbums[0].Title)
	})

	t.Run("User 2 should only see child2 album", func(t *testing.T) {
		returnedAlbums, err := actions.MyAlbums(db, user2, nil, nil, &boolTrue, &boolTrue, &boolFalse, nil)
		assert.NoError(t, err)

		assert.Len(t, returnedAlbums, 1)
		assert.Equal(t, "child2", returnedAlbums[0].Title)
	})
}

func TestMyAlbumsExcludesHidden(t *testing.T) {
	db := test_utils.DatabaseTest(t)
	boolTrue := true

	user, err := models.RegisterUser(db, "hide_myalbums_user", nil, false)
	assert.NoError(t, err)

	visible := models.Album{Title: "visible", Path: "/photos/mya_visible"}
	assert.NoError(t, db.Save(&visible).Error)
	hidden := models.Album{Title: "hidden", Path: "/photos/mya_hidden"}
	assert.NoError(t, db.Save(&hidden).Error)

	assert.NoError(t, db.Model(&user).Association("Albums").Append(&visible, &hidden))

	_, err = user.HideAlbum(db, hidden.ID, true)
	assert.NoError(t, err)

	t.Run("hidden album is excluded by default", func(t *testing.T) {
		albums, err := actions.MyAlbums(db, user, nil, nil, nil, &boolTrue, nil, nil)
		assert.NoError(t, err)
		titles := make([]string, len(albums))
		for i, a := range albums {
			titles[i] = a.Title
		}
		assert.Contains(t, titles, "visible")
		assert.NotContains(t, titles, "hidden")
	})

	t.Run("showHidden reveals it again", func(t *testing.T) {
		albums, err := actions.MyAlbums(db, user, nil, nil, nil, &boolTrue, nil, &boolTrue)
		assert.NoError(t, err)
		titles := make([]string, len(albums))
		for i, a := range albums {
			titles[i] = a.Title
		}
		assert.Contains(t, titles, "visible")
		assert.Contains(t, titles, "hidden")
	})
}

// TestMyAlbumsOnlyRootWithUnrelatedShare covers a real production bug: a
// user with a single true root album ("papa") who also receives an
// unrelated share of someone else's subfolder ("pc_media", whose real
// parent "meike_root" isn't in the user's own album set) used to have that
// share silently disappear - the single-root special case flattened
// "papa" to its own children without noticing "pc_media" wasn't one of
// them, so it never showed up anywhere in the onlyRoot listing.
func TestMyAlbumsOnlyRootWithUnrelatedShare(t *testing.T) {
	db := test_utils.DatabaseTest(t)
	boolTrue := true

	papa := models.Album{Title: "papa", Path: "/photos/papa"}
	assert.NoError(t, db.Save(&papa).Error)
	papaYear := models.Album{Title: "papa_year", Path: "/photos/papa/1999", ParentAlbumID: &papa.ID}
	assert.NoError(t, db.Save(&papaYear).Error)

	meikeRoot := models.Album{Title: "meike_root", Path: "/photos/meike"}
	assert.NoError(t, db.Save(&meikeRoot).Error)
	pcMedia := models.Album{Title: "pc_media", Path: "/photos/meike/pc_media", ParentAlbumID: &meikeRoot.ID}
	assert.NoError(t, db.Save(&pcMedia).Error)

	user, err := models.RegisterUser(db, "regina", nil, false)
	assert.NoError(t, err)

	// User has the whole "papa" tree, plus "pc_media" specifically - but
	// not "meike_root", its real parent.
	assert.NoError(t, db.Model(&user).Association("Albums").Append(&papa, &papaYear, &pcMedia))

	t.Run("papa and pc_media both show as their own top-level entries", func(t *testing.T) {
		albums, err := actions.MyAlbums(db, user, nil, nil, &boolTrue, &boolTrue, nil, nil)
		assert.NoError(t, err)

		titles := make([]string, len(albums))
		for i, a := range albums {
			titles[i] = a.Title
		}
		assert.ElementsMatch(t, []string{"papa", "pc_media"}, titles)
	})

	t.Run("removing access to papa leaves pc_media showing as itself", func(t *testing.T) {
		assert.NoError(t, db.Model(&user).Association("Albums").Delete(&papa, &papaYear))

		user.Albums = nil // force MyAlbums to reload the user's albums
		albums, err := actions.MyAlbums(db, user, nil, nil, &boolTrue, &boolTrue, nil, nil)
		assert.NoError(t, err)

		titles := make([]string, len(albums))
		for i, a := range albums {
			titles[i] = a.Title
		}
		assert.Equal(t, []string{"pc_media"}, titles)
	})
}
