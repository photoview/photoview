package resolvers

import (
	"context"
	"testing"

	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
)

func TestScanAlbumAuthorization(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	// The real queue starts a background worker on first use and walks the
	// filesystem; this test only cares who gets past the permission check.
	queued := make([]int, 0)
	origAddAlbumToQueue := addAlbumToQueue
	addAlbumToQueue = func(album *models.Album) error {
		queued = append(queued, album.ID)
		return nil
	}
	t.Cleanup(func() { addAlbumToQueue = origAddAlbumToQueue })

	album := models.Album{Title: "scan_album", Path: "/photos/scan_album"}
	assert.NoError(t, db.Save(&album).Error)

	owner, err := models.RegisterUser(db, "scan_owner", nil, false)
	assert.NoError(t, err)
	assert.NoError(t, db.Model(&owner).Association("Albums").Append(&album))

	stranger, err := models.RegisterUser(db, "scan_stranger", nil, false)
	assert.NoError(t, err)

	admin, err := models.RegisterUser(db, "scan_admin", nil, true)
	assert.NoError(t, err)

	r := &mutationResolver{Resolver: &Resolver{database: db}}

	t.Run("the album's owner may rescan it", func(t *testing.T) {
		queued = queued[:0]

		result, err := r.ScanAlbum(auth.AddUserToContext(context.Background(), owner), album.ID)
		assert.NoError(t, err)
		if assert.NotNil(t, result) {
			assert.True(t, result.Success)
		}
		assert.Equal(t, []int{album.ID}, queued)
	})

	t.Run("a user without access may not", func(t *testing.T) {
		queued = queued[:0]

		_, err := r.ScanAlbum(auth.AddUserToContext(context.Background(), stranger), album.ID)
		assert.Error(t, err)
		assert.Empty(t, queued, "a denied request must not reach the queue")
	})

	t.Run("an admin may rescan any album", func(t *testing.T) {
		queued = queued[:0]

		_, err := r.ScanAlbum(auth.AddUserToContext(context.Background(), admin), album.ID)
		assert.NoError(t, err)
		assert.Equal(t, []int{album.ID}, queued)
	})

	t.Run("an unauthenticated request is refused", func(t *testing.T) {
		_, err := r.ScanAlbum(context.Background(), album.ID)
		assert.Error(t, err)
	})
}
