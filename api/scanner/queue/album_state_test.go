package queue

import (
	"context"
	"sync"
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/graphql/notification"
	"github.com/photoview/photoview/api/test_utils"
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
	for i := 0; i < total; i++ {
		changed := i%3 == 0
		if changed {
			wantChanged++
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			state.CompleteMedia(context.Background(), nil, changed)
		}()
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

// TestCompleteMediaSkipAfterSuppressesCompletion verifies that a task
// submitted via ProcessMedia (skipAfter=true) never runs album-wide cleanup
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
