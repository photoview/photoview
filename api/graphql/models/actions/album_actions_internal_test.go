package actions

import (
	"context"
	"testing"
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAlbumAncestorsEndsOnACycle pins down why the ancestor walk carries no
// depth limit: UNION stops a cyclic parent chain by itself, as long as the
// recursion adds nothing that makes each lap's rows distinct. Nothing creates
// such a cycle today - parents come from the directory tree - but a bad row
// must not turn the breadcrumb query into a hang.
//
// It tests the walk alone because AlbumPath then checks ownership through
// GetParents, whose own recursive query does not end on a cycle yet.
func TestAlbumAncestorsEndsOnACycle(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	child := models.Album{
		Title:       "Child",
		Path:        "/cycle/child",
		ParentAlbum: &models.Album{Title: "Parent", Path: "/cycle"},
	}
	require.NoError(t, db.Save(&child).Error)
	require.NoError(t, db.Model(child.ParentAlbum).Update("parent_album_id", child.ID).Error)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ancestors, err := albumAncestors(db.WithContext(ctx), child.ID)
	require.NoError(t, err, "the walk must end on its own, not by the timeout")

	if assert.Len(t, ancestors, 1) {
		assert.Equal(t, "Parent", ancestors[0].Title)
	}
}

func TestAlbumAncestorsClosestFirst(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	// AlbumPath cuts the breadcrumb at the first ancestor the user cannot see,
	// counting from the album outward, so the order is part of the contract.
	leaf := models.Album{
		Title: "Leaf",
		Path:  "/order/a/b/leaf",
		ParentAlbum: &models.Album{
			Title: "B",
			Path:  "/order/a/b",
			ParentAlbum: &models.Album{
				Title:       "A",
				Path:        "/order/a",
				ParentAlbum: &models.Album{Title: "Root", Path: "/order"},
			},
		},
	}
	require.NoError(t, db.Save(&leaf).Error)

	ancestors, err := albumAncestors(db, leaf.ID)
	require.NoError(t, err)

	titles := make([]string, 0, len(ancestors))
	for _, ancestor := range ancestors {
		titles = append(titles, ancestor.Title)
	}

	assert.Equal(t, []string{"B", "A", "Root"}, titles)
}
