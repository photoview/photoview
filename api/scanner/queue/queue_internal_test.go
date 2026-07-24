package queue

import (
	"context"
	"fmt"
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/scanner_cache"
	"github.com/photoview/photoview/api/test_utils"
)

// TestQueueSubmittedAlbumsAreNotAbandonedOnClose checks that a burst of
// AlbumRequests submitted right before Close() are never abandoned.
// `incoming` is deliberately unbuffered: a send only returns once dispatch()
// has received it and, in that same select case (before it can next look at
// quit), absorbed it into pendingAlbums - so "the send returned" already
// means "this is in the queue", deterministically, not as a race dispatch()
// has to win. This is exercised directly against a hand-built Queue
// (bypassing the package-level singleton and Initialize) using empty temp
// directories, so no worker/exiftool process ever needs to start.
func TestQueueSubmittedAlbumsAreNotAbandonedOnClose(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	q := &Queue{
		ctx:      context.Background(),
		db:       db,
		incoming: make(chan any),
		quit:     make(chan struct{}),
		done:     make(chan struct{}),
	}
	q.max.Store(1)

	go q.dispatch()

	const numAlbums = 32
	cache := scanner_cache.MakeAlbumCache()
	for i := 0; i < numAlbums; i++ {
		album := &models.Album{Title: fmt.Sprintf("empty-album-%d", i), Path: t.TempDir()}
		if err := db.Create(album).Error; err != nil {
			t.Fatalf("create album %d: %v", i, err)
		}

		q.incoming <- AlbumRequest{Album: album, Cache: cache}
	}

	close(q.quit)
	<-q.done

	if len(q.pendingAlbums) != 0 {
		t.Errorf("pendingAlbums not empty after shutdown: %d abandoned", len(q.pendingAlbums))
	}

	select {
	case req := <-q.incoming:
		t.Errorf("found an unprocessed request left in incoming after shutdown: %+v", req)
	default:
	}
}
