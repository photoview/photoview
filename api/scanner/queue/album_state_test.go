package queue

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/graphql/notification"
	"github.com/photoview/photoview/api/scanner/face_detection"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/photoview/photoview/api/utils"
)

// TestCompleteMediaConcurrentCounting drives CompleteMedia from many
// goroutines at once (matching how several workers can finish tasks for the
// same album concurrently) and checks, under -race, that the shared
// remaining/changedCount bookkeeping ends up exact and that the "album
// complete" notification fires exactly once - not zero, not more than once -
// regardless of which goroutine happens to be the one that brings remaining
// to 0.
func TestCompleteMediaConcurrentCounting(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	album := &models.Album{Title: "concurrent completion test", Path: t.TempDir()}
	if err := db.Create(album).Error; err != nil {
		t.Fatalf("create album: %v", err)
	}

	const total = 64
	state := newAlbumState(db, album, nil, total)

	notifications := make(chan *models.Notification, total*2+10)
	listenerID := notification.RegisterListener(&models.User{}, notifications)
	t.Cleanup(func() {
		if err := notification.DeregisterListener(listenerID); err != nil {
			t.Errorf("deregister notification listener: %v", err)
		}
	})

	wantChanged := 0
	var wg sync.WaitGroup
	for i := range total {
		changed := i%3 == 0
		if changed {
			wantChanged++
		}
		wg.Go(func() {
			state.CompleteMedia(context.Background(), nil, changed)
		})
	}
	wg.Wait()

	if state.remaining != 0 {
		t.Errorf("remaining = %d, want 0", state.remaining)
	}
	if state.changedCount != wantChanged {
		t.Errorf("changedCount = %d, want %d", state.changedCount, wantChanged)
	}

	doneCount := 0
drain:
	for {
		select {
		case n := <-notifications:
			if n.Type == models.NotificationTypeMessage && n.Positive {
				doneCount++
			}
		default:
			break drain
		}
	}
	if doneCount != 1 {
		t.Errorf("got %d 'scan complete' notifications, want exactly 1", doneCount)
	}
}

// TestCleanupStaleMediaDeletesGoneMedia checks the branch cleanupStaleMedia
// takes when it actually finds stale media: rows outside allMediaIDs must be
// deleted from the database along with their cache directories, while media
// that was found stays untouched. It also exercises the
// face_detection.GlobalFaceDetector reload that only runs on this branch.
func TestCleanupStaleMediaDeletesGoneMedia(t *testing.T) {
	test_utils.FilesystemTest(t)
	db := test_utils.DatabaseTest(t)

	if err := face_detection.InitializeFaceDetector(db); err != nil {
		t.Fatalf("initialize face detector: %v", err)
	}

	album := &models.Album{Title: "stale media test", Path: t.TempDir()}
	if err := db.Create(album).Error; err != nil {
		t.Fatalf("create album: %v", err)
	}

	makeMedia := func(name string) *models.Media {
		m := &models.Media{
			Title:    name,
			Path:     filepath.Join(album.Path, name),
			AlbumID:  album.ID,
			Type:     models.MediaTypePhoto,
			DateShot: time.Now(),
		}
		if err := db.Create(m).Error; err != nil {
			t.Fatalf("create media %s: %v", name, err)
		}
		return m
	}

	keep := makeMedia("keep.jpg")
	gone1 := makeMedia("gone1.jpg")
	gone2 := makeMedia("gone2.jpg")

	var goneCacheDirs []string
	for _, m := range []*models.Media{keep, gone1, gone2} {
		cachePath, err := utils.CachePathForMedia(album.ID, m.ID)
		if err != nil {
			t.Fatalf("cache path for media %d: %v", m.ID, err)
		}
		if err := os.WriteFile(filepath.Join(cachePath, "thumbnail.jpg"), []byte("x"), 0o644); err != nil {
			t.Fatalf("write cache file: %v", err)
		}
		if m.ID != keep.ID {
			goneCacheDirs = append(goneCacheDirs, cachePath)
		}
	}

	state := newAlbumState(db, album, nil, 1)
	state.allMediaIDs = []int{keep.ID}

	if err := state.cleanupStaleMedia(context.Background()); err != nil {
		t.Fatalf("cleanupStaleMedia() error: %v", err)
	}

	var remaining []models.Media
	if err := db.Where("album_id = ?", album.ID).Find(&remaining).Error; err != nil {
		t.Fatalf("query remaining media: %v", err)
	}
	if len(remaining) != 1 || remaining[0].ID != keep.ID {
		t.Errorf("remaining media = %+v, want only media %d (%q)", remaining, keep.ID, keep.Title)
	}

	for _, dir := range goneCacheDirs {
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			t.Errorf("stale cache dir %s still exists, want it removed", dir)
		}
	}

	keepCachePath, err := utils.CachePathForMedia(album.ID, keep.ID)
	if err != nil {
		t.Fatalf("cache path for kept media: %v", err)
	}
	if _, err := os.Stat(filepath.Join(keepCachePath, "thumbnail.jpg")); err != nil {
		t.Errorf("kept media's cache file was removed: %v", err)
	}
}

// TestCompleteMediaSkipAfterSuppressesCompletion verifies that a task
// submitted via queue.SubmitMedia (skipAfter=true) never runs album-wide cleanup
// or broadcasts the "scan complete" notification, even once remaining
// reaches 0.
func TestCompleteMediaSkipAfterSuppressesCompletion(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	album := &models.Album{Title: "skip after test", Path: t.TempDir()}
	if err := db.Create(album).Error; err != nil {
		t.Fatalf("create album: %v", err)
	}

	state := newAlbumState(db, album, nil, 1)
	state.skipAfter = true

	notifications := make(chan *models.Notification, 10)
	listenerID := notification.RegisterListener(&models.User{}, notifications)
	t.Cleanup(func() {
		if err := notification.DeregisterListener(listenerID); err != nil {
			t.Errorf("deregister notification listener: %v", err)
		}
	})

	state.CompleteMedia(context.Background(), nil, true)

	if state.remaining != 0 {
		t.Errorf("remaining = %d, want 0", state.remaining)
	}

drain:
	for {
		select {
		case n := <-notifications:
			if n.Type == models.NotificationTypeMessage && n.Positive {
				t.Errorf("got a 'scan complete' notification, want none since skipAfter is set")
			}
		default:
			break drain
		}
	}
}
