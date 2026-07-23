package models_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const concurrentUpsertCallers = 32

// TestFindOrCreateAlbumConcurrent drives FindOrCreateAlbum from many
// goroutines at once for the same path, matching two overlapping album
// scans (e.g. two users sharing an album tree) racing to create the same
// row. All callers must converge on exactly one row with the same ID.
func TestFindOrCreateAlbumConcurrent(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	const albumPath = "/concurrent/album/path"

	ids := make([]int, concurrentUpsertCallers)
	var wg sync.WaitGroup
	for i := 0; i < concurrentUpsertCallers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			album := &models.Album{
				Title: "concurrent album",
				Path:  albumPath,
			}
			_, err := models.FindOrCreateAlbum(db, album)
			assert.NoError(t, err)
			ids[i] = album.ID
		}(i)
	}
	wg.Wait()

	for i, id := range ids {
		assert.NotZero(t, id, "caller %d got a zero ID", i)
		assert.Equal(t, ids[0], id, "caller %d converged on a different album than caller 0", i)
	}

	var count int64
	require.NoError(t, db.Model(&models.Album{}).Where("path_hash = ?", models.MD5Hash(albumPath)).Count(&count).Error)
	assert.EqualValues(t, 1, count, "expected exactly one album row for the shared path")
}

// TestFindOrCreateMediaConcurrent mirrors TestFindOrCreateAlbumConcurrent for
// Media: many goroutines racing to scan the same media path must converge on
// exactly one row.
func TestFindOrCreateMediaConcurrent(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	album := &models.Album{Title: "media upsert test album", Path: "/concurrent/media/album"}
	require.NoError(t, db.Create(album).Error)

	const mediaPath = "/concurrent/media/album/photo.jpg"

	ids := make([]int, concurrentUpsertCallers)
	var wg sync.WaitGroup
	for i := 0; i < concurrentUpsertCallers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			media := &models.Media{
				Title:    "photo.jpg",
				Path:     mediaPath,
				AlbumID:  album.ID,
				Type:     models.MediaTypePhoto,
				DateShot: time.Now(),
			}
			_, err := models.FindOrCreateMedia(db, media)
			assert.NoError(t, err)
			ids[i] = media.ID
		}(i)
	}
	wg.Wait()

	for i, id := range ids {
		assert.NotZero(t, id, "caller %d got a zero ID", i)
		assert.Equal(t, ids[0], id, "caller %d converged on a different media than caller 0", i)
	}

	var count int64
	require.NoError(t, db.Model(&models.Media{}).Where("path_hash = ?", models.MD5Hash(mediaPath)).Count(&count).Error)
	assert.EqualValues(t, 1, count, "expected exactly one media row for the shared path")
}

// TestUpsertMediaURLConcurrent drives UpsertMediaURL from many goroutines at
// once for the same (media_id, purpose) pair - matching two workers that
// independently decided the same cache file needed to be (re)generated.
// Exactly one row must exist afterwards (whichever caller's write landed
// last), never duplicates.
func TestUpsertMediaURLConcurrent(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	album := &models.Album{Title: "media url upsert test album", Path: "/concurrent/media_url/album"}
	require.NoError(t, db.Create(album).Error)

	media := &models.Media{
		Title:    "video.mp4",
		Path:     "/concurrent/media_url/album/video.mp4",
		AlbumID:  album.ID,
		Type:     models.MediaTypeVideo,
		DateShot: time.Now(),
	}
	require.NoError(t, db.Create(media).Error)

	var wg sync.WaitGroup
	for i := 0; i < concurrentUpsertCallers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			url := &models.MediaURL{
				MediaID:     media.ID,
				MediaName:   fmt.Sprintf("video-web-%d.mp4", i),
				Width:       1280,
				Height:      720,
				Purpose:     models.VideoWeb,
				ContentType: "video/mp4",
				FileSize:    int64(i),
			}
			assert.NoError(t, models.UpsertMediaURL(db, url))
		}(i)
	}
	wg.Wait()

	var urls []models.MediaURL
	require.NoError(t, db.Where("media_id = ? AND purpose = ?", media.ID, models.VideoWeb).Find(&urls).Error)
	assert.Len(t, urls, 1, "expected exactly one media url row for the shared (media_id, purpose) pair")
}
