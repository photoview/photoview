package resolvers

import (
	"context"
	"errors"
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

func TestScanAlbumReportsFailures(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	admin, err := models.RegisterUser(db, "scan_fail_admin", nil, true)
	assert.NoError(t, err)
	ctx := auth.AddUserToContext(context.Background(), admin)

	r := &mutationResolver{Resolver: &Resolver{database: db}}

	t.Run("an album that does not exist", func(t *testing.T) {
		result, err := r.ScanAlbum(ctx, 987654)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("a queue that cannot take the album", func(t *testing.T) {
		// A directory that cannot be walked must come back as an error, not
		// as a "Scanner started" the user would wait on for nothing.
		origAddAlbumToQueue := addAlbumToQueue
		addAlbumToQueue = func(*models.Album) error { return errors.New("album directory does not exist") }
		t.Cleanup(func() { addAlbumToQueue = origAddAlbumToQueue })

		album := models.Album{Title: "scan_fail_album", Path: "/photos/scan_fail_album"}
		assert.NoError(t, db.Save(&album).Error)

		result, err := r.ScanAlbum(ctx, album.ID)
		assert.ErrorContains(t, err, "album directory does not exist")
		assert.Nil(t, result)
	})
}
