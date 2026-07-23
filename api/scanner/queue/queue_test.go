package queue_test

import (
	"context"
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/face_detection"
	"github.com/photoview/photoview/api/scanner/queue"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	test_utils.IntegrationTestRun(m)
}

// TestProcessMediaDoesNotDeleteSiblingMedia guards against the albumState.skipAfter
// bug: ProcessMedia builds a throwaway, single-media albumState so it must not run
// album-wide cleanup, which would otherwise treat every *other* media in the album as
// stale and delete it.
func TestProcessMediaDoesNotDeleteSiblingMedia(t *testing.T) {
	test_utils.FilesystemTest(t)
	db := test_utils.DatabaseTest(t)

	pass := "1234"
	user, err := models.RegisterUser(db, "queue_test_user", &pass, true)
	require.NoError(t, err)

	rootAlbum := models.Album{
		Title: "root album",
		Path:  test_utils.PathFromAPIRoot("scanner", "test_media", "orient"),
	}
	require.NoError(t, db.Save(&rootAlbum).Error)
	require.NoError(t, db.Model(user).Association("Albums").Append(&rootAlbum))

	require.NoError(t, face_detection.InitializeFaceDetector(db))

	// Full scan of the album first, so it has several sibling media in the database.
	require.NoError(t, queue.Initialize(context.Background(), db))
	require.NoError(t, queue.AddUser(user))
	queue.Close()

	var before []models.Media
	require.NoError(t, db.Where("album_id = ?", rootAlbum.ID).Find(&before).Error)
	require.NotEmpty(t, before, "expected the initial scan to find media in test_media/orient")
	require.Greater(t, len(before), 1, "need at least 2 media to prove siblings survive")

	target := before[0]

	// Re-initialize the queue (Close() above stopped the previous dispatcher) and
	// reprocess a single file through queue.ProcessMedia, same as routes/videos.go does.
	require.NoError(t, queue.Initialize(context.Background(), db))
	require.NoError(t, queue.ProcessMedia(context.Background(), db, &target))
	queue.Close()

	var after []models.Media
	require.NoError(t, db.Where("album_id = ?", rootAlbum.ID).Find(&after).Error)
	assert.Len(t, after, len(before), "ProcessMedia must not delete sibling media in the same album")

	afterIDs := make(map[int]bool, len(after))
	for _, m := range after {
		afterIDs[m.ID] = true
	}
	for _, m := range before {
		assert.Truef(t, afterIDs[m.ID], "media %q (id=%d) should still exist after ProcessMedia", m.Title, m.ID)
	}
}
