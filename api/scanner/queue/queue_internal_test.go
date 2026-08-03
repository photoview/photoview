package queue

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/face_detection"
	"github.com/photoview/photoview/api/scanner/scanner_cache"
	"github.com/photoview/photoview/api/test_utils"
)

func TestMain(m *testing.M) {
	test_utils.IntegrationTestRun(m)
}

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
		ctx:            context.Background(),
		db:             db,
		incoming:       make(chan any),
		quit:           make(chan struct{}),
		done:           make(chan struct{}),
		scanningAlbums: make(map[int]struct{}),
	}
	q.max.Store(1)

	go q.dispatch()

	const numAlbums = 32
	cache := scanner_cache.MakeAlbumCache()
	for i := range numAlbums {
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

// TestChangeConcurrentWorkers checks that it updates the live singleton's
// worker pool ceiling.
func TestChangeConcurrentWorkers(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	if err := Initialize(context.Background(), db); err != nil {
		t.Fatalf("Initialize() error: %v", err)
	}
	defer Close()

	ChangeConcurrentWorkers(7)

	if got := globalQueue.max.Load(); got != 7 {
		t.Errorf("globalQueue.max = %d, want 7", got)
	}
}

// TestUninitializedQueueCallsAreSafe checks that every exported entry point
// handles a nil globalQueue (i.e. called before Initialize, or after Close)
// without panicking.
func TestUninitializedQueueCallsAreSafe(t *testing.T) {
	prev := globalQueue
	globalQueue = nil
	defer func() { globalQueue = prev }()

	Close() // must not panic

	ChangeConcurrentWorkers(4) // must not panic

	if Initialized() {
		t.Errorf("Initialized() = true, want false when globalQueue is nil")
	}

	if err := SubmitAlbum(&models.Album{}, nil); err == nil {
		t.Errorf("SubmitAlbum() error = nil, want an error when the queue is not initialized")
	}

	if err := SubmitMedia(context.Background(), nil, &models.Album{}, nil, ""); err == nil {
		t.Errorf("SubmitMedia() error = nil, want an error when the queue is not initialized")
	}
}

// TestDispatchLogsExpandAlbumErrorAndContinues checks that dispatch()
// recovers from a bad AlbumRequest (a path that no longer exists on disk)
// instead of wedging - a later, valid request must still get scanned.
func TestDispatchLogsExpandAlbumErrorAndContinues(t *testing.T) {
	test_utils.FilesystemTest(t)
	db := test_utils.DatabaseTest(t)

	if err := face_detection.InitializeFaceDetector(db); err != nil {
		t.Fatalf("initialize face detector: %v", err)
	}

	// Both albums are created before the dispatcher starts, so nothing
	// fallible runs between "goroutine started" and "goroutine signaled to
	// quit" below - a t.Fatalf here can't leak the dispatcher goroutine.
	badAlbum := &models.Album{Title: "missing directory", Path: filepath.Join(t.TempDir(), "does-not-exist")}
	if err := db.Create(badAlbum).Error; err != nil {
		t.Fatalf("create bad album: %v", err)
	}

	goodDir := t.TempDir()
	copyFixtureFile(t, "photo/plain.jpg", filepath.Join(goodDir, "plain.jpg"))
	goodAlbum := &models.Album{Title: "scannable album", Path: goodDir}
	if err := db.Create(goodAlbum).Error; err != nil {
		t.Fatalf("create good album: %v", err)
	}

	q := &Queue{
		ctx:            context.Background(),
		db:             db,
		incoming:       make(chan any),
		quit:           make(chan struct{}),
		done:           make(chan struct{}),
		scanningAlbums: make(map[int]struct{}),
	}
	q.max.Store(1)

	go q.dispatch()

	q.incoming <- AlbumRequest{Album: badAlbum, Cache: scanner_cache.MakeAlbumCache()}
	q.incoming <- AlbumRequest{Album: goodAlbum, Cache: scanner_cache.MakeAlbumCache()}

	close(q.quit)
	<-q.done

	var count int64
	if err := db.Model(&models.Media{}).Where("album_id = ?", goodAlbum.ID).Count(&count).Error; err != nil {
		t.Fatalf("count good album media: %v", err)
	}
	if count == 0 {
		t.Errorf("expected the good album to still be scanned after the bad one errored, found 0 media")
	}
}

// TestCompleteMediaLastTaskDoesNotDeadlockDispatchTeardown reproduces a
// deadlock: dispatch() tears down its worker pool (close(taskChan) then
// wg.Wait()) the instant pendingTasks empties, which happens the moment the
// last task is handed to a worker - not when that worker actually finishes
// it. If CompleteMedia's isLast branch then tries to report the album done
// back to dispatch() over q.incoming, nothing is reading q.incoming anymore
// (dispatch() is parked in wg.Wait()), so the worker blocks forever trying
// to send and dispatch() blocks forever waiting for it to exit - neither
// side can proceed.
//
// If this test ever fails by timing out, its own t.Cleanup below
// (close(q.quit); <-q.done) will hang too, since a genuinely deadlocked
// dispatch() goroutine can never reach that select - nothing can force it to
// unwind from outside. That's an inherent limit of testing a real deadlock
// in-process: go test's own overall -timeout is what ultimately kills the
// binary in that case, not this test.
func TestCompleteMediaLastTaskDoesNotDeadlockDispatchTeardown(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	album := &models.Album{Title: "deadlock repro", Path: t.TempDir()}
	if err := db.Create(album).Error; err != nil {
		t.Fatalf("create album: %v", err)
	}

	q := &Queue{
		ctx:            context.Background(),
		db:             db,
		incoming:       make(chan any),
		quit:           make(chan struct{}),
		done:           make(chan struct{}),
		scanningAlbums: make(map[int]struct{}),
	}
	q.max.Store(1)

	go q.dispatch()
	t.Cleanup(func() {
		close(q.quit)
		<-q.done
	})

	cache := scanner_cache.MakeAlbumCache()
	state := newAlbumState(db, album, cache, 1)
	state.markDone = q.markAlbumDone

	taskDone := make(chan struct{})
	// The path doesn't need to exist or be a real image: gather() fails
	// os.Stat-ing it, but processMedia's defer calls CompleteMedia
	// regardless of how the task ended - which is all this test needs to
	// reach the isLast branch.
	tsk := newTask(album, cache, state, filepath.Join(album.Path, "does-not-exist.jpg"), taskDone)

	q.incoming <- tsk

	select {
	case <-taskDone:
	case <-time.After(5 * time.Second):
		t.Fatal("processMedia never returned for the album's last task - dispatch() likely deadlocked tearing down its worker pool while the worker waited to report the album done")
	}
}
