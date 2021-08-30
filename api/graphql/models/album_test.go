package models_test

import (
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
)

func TestAlbumGetChildrenAndParents(t *testing.T) {
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
		{
			Title: "not_child",
			Path:  "/videos",
		},
	}

	if !assert.NoError(t, db.Save(&children).Error) {
		return
	}

	sub_child := models.Album{
		Title:         "subchild",
		Path:          "/photos/child1/subchild",
		ParentAlbumID: &children[0].ID,
	}

	if !assert.NoError(t, db.Save(&sub_child).Error) {
		return
	}

	verifyResult := func(t *testing.T, expected_albums []*models.Album, result []*models.Album) {
		assert.Equal(t, len(expected_albums), len(result))

		for _, expected := range expected_albums {
			found_expected := false
			for _, item := range result {
				if item.Title == expected.Title && item.Path == expected.Path {
					found_expected = true
					break
				}
			}
			if !found_expected {
				assert.Failf(t, "albums did not match", "expected to find item: %v", expected)
			}
		}
	}

	t.Run("Album get children", func(t *testing.T) {
		root_children, err := rootAlbum.GetChildren(db, nil)
		if !assert.NoError(t, err) {
			return
		}

		expected_children := []*models.Album{
			{
				Title: "root",
				Path:  "/photos",
			},
			{
				Title: "child1",
				Path:  "/photos/child1",
			},
			{
				Title: "child2",
				Path:  "/photos/child2",
			},
			{
				Title: "subchild",
				Path:  "/photos/child1/subchild",
			},
		}

		verifyResult(t, expected_children, root_children)
	})

	t.Run("Album get parents", func(t *testing.T) {
		parents, err := sub_child.GetParents(db, nil)
		if !assert.NoError(t, err) {
			return
		}

		expected_parents := []*models.Album{
			{
				Title: "root",
				Path:  "/photos",
			},
			{
				Title: "child1",
				Path:  "/photos/child1",
			},
			{
				Title: "subchild",
				Path:  "/photos/child1/subchild",
			},
		}

		verifyResult(t, expected_parents, parents)
	})

}
