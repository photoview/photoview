package models_test

import (
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
)

func TestAlbumGetChildren(t *testing.T) {
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

	assert.Equal(t, len(expected_children), len(root_children))

	for _, expected := range expected_children {
		found_expected := false
		for _, item := range root_children {
			if item.Title == expected.Title && item.Path == expected.Path {
				found_expected = true
				break
			}
		}
		if !found_expected {
			assert.Failf(t, "root children did not match", "expected to find item: %v", expected)
		}
	}

}
