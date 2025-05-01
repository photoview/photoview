package models_test

import (
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
)

func TestAlbumGetChildrenAndParents(t *testing.T) {
	const photosPath = "/photos"
	const photosChild1Path = "/photos/child1"
	const photosChild1SubchildPath = "/photos/child1/subchild"
	db := test_utils.DatabaseTest(t)

	rootAlbum := models.Album{
		Title: "root",
		Path:  photosPath,
	}

	if !assert.NoError(t, db.Save(&rootAlbum).Error) {
		return
	}

	children := []models.Album{
		{
			Title:         "child1",
			Path:          photosChild1Path,
			ParentAlbumID: &rootAlbum.ID,
		},
		{
			Title:         "child2",
			Path:          "/photos/child2",
			ParentAlbumID: &rootAlbum.ID,
		},
		{
			Title: "not_child",
			Path:  "/videos",
		},
	}

	if !assert.NoError(t, db.Save(&children).Error) {
		return
	}

	subChild := models.Album{
		Title:         "subchild",
		Path:          photosChild1SubchildPath,
		ParentAlbumID: &children[0].ID,
	}

	if !assert.NoError(t, db.Save(&subChild).Error) {
		return
	}

	verifyResult := func(t *testing.T, expectedAlbums []*models.Album, result []*models.Album) {
		assert.Equal(t, len(expectedAlbums), len(result))

		for _, expected := range expectedAlbums {
			foundExpected := false
			for _, item := range result {
				if item.Title == expected.Title && item.Path == expected.Path {
					foundExpected = true
					break
				}
			}
			if !foundExpected {
				assert.Failf(t, "albums did not match", "expected to find item: %v", expected)
			}
		}
	}

	t.Run("Album get children", func(t *testing.T) {
		rootChildren, err := rootAlbum.GetChildren(db, nil)
		if !assert.NoError(t, err) {
			return
		}

		expectedChildren := []*models.Album{
			{
				Title: "root",
				Path:  photosPath,
			},
			{
				Title: "child1",
				Path:  photosChild1Path,
			},
			{
				Title: "child2",
				Path:  "/photos/child2",
			},
			{
				Title: "subchild",
				Path:  photosChild1SubchildPath,
			},
		}

		verifyResult(t, expectedChildren, rootChildren)
	})

	t.Run("Album get parents", func(t *testing.T) {
		parents, err := subChild.GetParents(db, nil)
		if !assert.NoError(t, err) {
			return
		}

		expectedParents := []*models.Album{
			{
				Title: "root",
				Path:  photosPath,
			},
			{
				Title: "child1",
				Path:  photosChild1Path,
			},
			{
				Title: "subchild",
				Path:  photosChild1SubchildPath,
			},
		}

		verifyResult(t, expectedParents, parents)
	})

}

func TestAlbumThumbnail(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	mediaAlbum := models.Album{
		Title: "Media album",
		Path:  "/media_album",
	}
	if !assert.NoError(t, db.Save(&mediaAlbum).Error) {
		return
	}

	media := models.Media{
		Path:    "thumb.jpg",
		AlbumID: mediaAlbum.ID,
	}
	if !assert.NoError(t, db.Save(&media).Error) {
		return
	}

	t.Run("Thumbnail from CoverID", func(t *testing.T) {
		album := models.Album{
			Title:   "Album with cover",
			Path:    "/cover_album",
			CoverID: &media.ID,
		}
		if !assert.NoError(t, db.Save(&album).Error) {
			return
		}

		result, err := album.Thumbnail(db)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, media.ID, result.ID)
	})

	t.Run("Thumbnail from child media", func(t *testing.T) {
		parentAlbum := models.Album{
			Title: "Parent album",
			Path:  "/parent",
		}
		if !assert.NoError(t, db.Save(&parentAlbum).Error) {
			return
		}

		childAlbum := models.Album{
			Title:         "Child album",
			Path:          "/parent/child",
			ParentAlbumID: &parentAlbum.ID,
		}
		if !assert.NoError(t, db.Save(&childAlbum).Error) {
			return
		}

		childMedia := models.Media{
			Path:    "child_media.jpg",
			AlbumID: childAlbum.ID,
		}
		if !assert.NoError(t, db.Save(&childMedia).Error) {
			return
		}

		result, err := parentAlbum.Thumbnail(db)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, childMedia.ID, result.ID)
	})

	t.Run("Empty album with no media", func(t *testing.T) {
		emptyAlbum := models.Album{
			Title: "Empty album",
			Path:  "/empty",
		}
		if !assert.NoError(t, db.Save(&emptyAlbum).Error) {
			return
		}

		result, err := emptyAlbum.Thumbnail(db)
		assert.NoError(t, err)
		assert.NotNil(t, result, "Album with no media should return an empty Media object")
		assert.Equal(t, 0, result.ID, "Empty album thumbnail should have ID=0")
	})

	t.Run("Thumbnail from grandchild media", func(t *testing.T) {
		// Create grandparent-parent-child relationship with media only in child
		grandparentAlbum := models.Album{
			Title: "Grandparent",
			Path:  "/grandparent",
		}
		if !assert.NoError(t, db.Save(&grandparentAlbum).Error) {
			return
		}

		parentAlbum := models.Album{
			Title:         "Parent",
			Path:          "/grandparent/parent",
			ParentAlbumID: &grandparentAlbum.ID,
		}
		if !assert.NoError(t, db.Save(&parentAlbum).Error) {
			return
		}

		childAlbum := models.Album{
			Title:         "Child",
			Path:          "/grandparent/parent/child",
			ParentAlbumID: &parentAlbum.ID,
		}
		if !assert.NoError(t, db.Save(&childAlbum).Error) {
			return
		}

		childMedia := models.Media{
			Path:    "deep_media.jpg",
			AlbumID: childAlbum.ID,
		}
		if !assert.NoError(t, db.Save(&childMedia).Error) {
			return
		}

		result, err := grandparentAlbum.Thumbnail(db)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, childMedia.ID, result.ID)
	})
}
