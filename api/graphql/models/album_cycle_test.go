package models_test

import (
	"context"
	"testing"
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// saveAlbumCycle stores two albums whose parents point at each other. The
// scanner never creates this, but a hand-edited or corrupted database can hold
// it, and a recursive album query must still finish on it.
func saveAlbumCycle(t *testing.T, db *gorm.DB) (first, second *models.Album) {
	first = &models.Album{Title: "first", Path: "/cycle/first"}
	require.NoError(t, db.Save(first).Error)

	second = &models.Album{Title: "second", Path: "/cycle/second", ParentAlbumID: &first.ID}
	require.NoError(t, db.Save(second).Error)

	require.NoError(t, db.Model(first).Update("parent_album_id", second.ID).Error)

	return first, second
}

// withQueryTimeout bounds every query, so a recursive query that never ends
// fails the test instead of hanging the suite.
func withQueryTimeout(t *testing.T, db *gorm.DB) *gorm.DB {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	return db.WithContext(ctx)
}

func albumIDs(albums []*models.Album) []int {
	ids := make([]int, len(albums))
	for i, album := range albums {
		ids[i] = album.ID
	}

	return ids
}

func TestAlbumQueriesFinishOnCyclicParents(t *testing.T) {
	db := test_utils.DatabaseTest(t)
	first, second := saveAlbumCycle(t, db)

	t.Run("GetParents", func(t *testing.T) {
		parents, err := first.GetParents(withQueryTimeout(t, db), nil)
		require.NoError(t, err)
		assert.ElementsMatch(t, []int{first.ID, second.ID}, albumIDs(parents))
	})

	t.Run("GetChildren", func(t *testing.T) {
		children, err := first.GetChildren(withQueryTimeout(t, db), nil)
		require.NoError(t, err)
		assert.ElementsMatch(t, []int{first.ID, second.ID}, albumIDs(children))
	})

	t.Run("GetChildrenFromAlbums", func(t *testing.T) {
		children, err := models.GetChildrenFromAlbums(withQueryTimeout(t, db), nil, []int{first.ID, second.ID})
		require.NoError(t, err)
		assert.ElementsMatch(t, []int{first.ID, second.ID}, albumIDs(children))
	})

	t.Run("OwnsAlbum", func(t *testing.T) {
		user, err := models.RegisterUser(db, "stranger", nil, false)
		require.NoError(t, err)

		owns, err := user.OwnsAlbum(withQueryTimeout(t, db), first)
		require.NoError(t, err)
		assert.False(t, owns)
	})

	t.Run("Thumbnail", func(t *testing.T) {
		media := models.Media{Path: "/cycle/second/photo.jpg", AlbumID: second.ID}
		require.NoError(t, db.Save(&media).Error)

		thumbnail, err := first.Thumbnail(withQueryTimeout(t, db))
		require.NoError(t, err)
		require.NotNil(t, thumbnail)
		assert.Equal(t, media.ID, thumbnail.ID)
	})
}
