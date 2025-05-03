package models_test

import (
	"testing"

	"github.com/kkovaletp/photoview/api/graphql/models"
	"github.com/kkovaletp/photoview/api/test_utils"
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
