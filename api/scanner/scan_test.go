package scanner_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner"
	"github.com/photoview/photoview/api/scanner/face_detection"
	"github.com/photoview/photoview/api/scanner/queue"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// copyFixtureFile copies a single file from api/scanner/test_media/library
// to dst, so tests can freely process their own copy without touching the
// checked-in fixture.
func copyFixtureFile(t *testing.T, name string, dst string) {
	t.Helper()

	src := test_utils.PathFromAPIRoot("scanner", "test_media", "library", name)
	data, err := os.ReadFile(src)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(dst, data, 0o644))
}

// TestAddAllQueuesEveryUser checks that AddAll finds and scans every user's
// albums, not just one - the loop in AddAll is otherwise indistinguishable
// from AddUser in test coverage.
func TestAddAllQueuesEveryUser(t *testing.T) {
	test_utils.FilesystemTest(t)
	db := test_utils.DatabaseTest(t)

	pass := "1234"
	user1, err := models.RegisterUser(db, "add_all_user_1", &pass, true)
	require.NoError(t, err)
	user2, err := models.RegisterUser(db, "add_all_user_2", &pass, true)
	require.NoError(t, err)

	album1Dir := t.TempDir()
	copyFixtureFile(t, "mount_merapi_volcano_indonesia.jpg", filepath.Join(album1Dir, "photo.jpg"))
	album1 := &models.Album{Title: "add-all album 1", Path: album1Dir}
	require.NoError(t, db.Save(album1).Error)
	require.NoError(t, db.Model(user1).Association("Albums").Append(album1))

	album2Dir := t.TempDir()
	copyFixtureFile(t, "buttercup_close_summer_yellow.jpg", filepath.Join(album2Dir, "photo.jpg"))
	album2 := &models.Album{Title: "add-all album 2", Path: album2Dir}
	require.NoError(t, db.Save(album2).Error)
	require.NoError(t, db.Model(user2).Association("Albums").Append(album2))

	require.NoError(t, face_detection.InitializeFaceDetector(db))

	require.NoError(t, queue.Initialize(context.Background(), db))
	// Close() unconditionally, even if AddAll() failed, so the dispatcher
	// goroutine it started is never left running past this test.
	addAllErr := scanner.AddAll(db)
	queue.Close()
	require.NoError(t, addAllErr)

	var count1, count2 int64
	require.NoError(t, db.Model(&models.Media{}).Where("album_id = ?", album1.ID).Count(&count1).Error)
	require.NoError(t, db.Model(&models.Media{}).Where("album_id = ?", album2.ID).Count(&count2).Error)
	assert.NotZero(t, count1, "expected AddAll to have scanned user1's album, found 0 media")
	assert.NotZero(t, count2, "expected AddAll to have scanned user2's album, found 0 media")
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
	require.NoError(t, scanner.AddUser(db, user))
	queue.Close()

	var before []models.Media
	require.NoError(t, db.Where("album_id = ?", rootAlbum.ID).Find(&before).Error)
	require.NotEmpty(t, before, "expected the initial scan to find media in test_media/orient")
	require.Greater(t, len(before), 1, "need at least 2 media to prove siblings survive")

	target := before[0]

	// Re-initialize the queue (Close() above stopped the previous dispatcher) and
	// reprocess a single file through scanner.ProcessMedia, same as routes/videos.go does.
	require.NoError(t, queue.Initialize(context.Background(), db))
	require.NoError(t, scanner.ProcessMedia(context.Background(), db, &target))
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
