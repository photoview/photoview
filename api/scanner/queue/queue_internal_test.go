package queue

import (
	"context"
	"fmt"
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/scanner_cache"
	"github.com/photoview/photoview/api/test_utils"
)

// TestQueueDrainsBufferedAlbumsBeforeShutdown is a regression test for a bug
// found while building this package: the dispatcher's idle-wait select has
// both an `incoming` and a `quit` case, and Go picks randomly between ready
// cases - so if Close() is called right after a burst of AlbumRequests were
// pushed, the dispatcher could pick the quit case and return before ever
// looking at the (already fully buffered) requests sitting in `incoming`,
// silently abandoning them. dispatch() must drain `incoming` before actually
// stopping. This is exercised directly against a hand-built Queue (bypassing
// the package-level singleton and Initialize) using empty temp directories,
// so no worker/exiftool process ever needs to start.
func TestQueueDrainsBufferedAlbumsBeforeShutdown(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	q := &Queue{
		ctx:      context.Background(),
		db:       db,
		incoming: make(chan any, 64),
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

		// Push directly onto the channel Close() will race against, then
		// immediately close - this is the exact race window the bug lived in.
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
