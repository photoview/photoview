package queue

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/face_detection"
	"github.com/photoview/photoview/api/scanner/scanner_cache"
	"github.com/photoview/photoview/api/test_utils"
	"gorm.io/gorm"
)

func TestMain(m *testing.M) {
	test_utils.IntegrationTestRun(m)
}

// TestQueueSubmittedAlbumsAreNotAbandonedOnClose checks that a burst of
// AlbumRequests submitted right before Close() are never abandoned.
// `albums` is deliberately unbuffered: a send only returns once dispatch()
// has received it and, in that same select case (before it can next look at
// quit), absorbed it into pendingAlbums - so "the send returned" already
// means "this is in the queue", deterministically, not as a race dispatch()
// has to win. This is exercised directly against a hand-built Queue
// (bypassing the package-level singleton and Initialize) using empty temp
// directories, so no worker/exiftool process ever needs to start.
func TestQueueSubmittedAlbumsAreNotAbandonedOnClose(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	q := newQueue(context.Background(), db, 1)

	go q.dispatch()

	const numAlbums = 32
	cache := scanner_cache.MakeAlbumCache()
	for i := range numAlbums {
		album := &models.Album{Title: fmt.Sprintf("empty-album-%d", i), Path: t.TempDir()}
		if err := db.Create(album).Error; err != nil {
			t.Fatalf("create album %d: %v", i, err)
		}

		q.albums <- AlbumRequest{Album: album, Cache: cache}
	}

	close(q.quit)
	<-q.done

	if len(q.pendingAlbums) != 0 {
		t.Errorf("pendingAlbums not empty after shutdown: %d abandoned", len(q.pendingAlbums))
	}

	select {
	case req := <-q.albums:
		t.Errorf("found an unprocessed request left in albums after shutdown: %+v", req)
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

	q := newQueue(context.Background(), db, 1)

	go q.dispatch()

	q.albums <- AlbumRequest{Album: badAlbum, Cache: scanner_cache.MakeAlbumCache()}
	q.albums <- AlbumRequest{Album: goodAlbum, Cache: scanner_cache.MakeAlbumCache()}

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

// TestCompleteMediaLastTaskDoesNotDeadlockDispatch checks that a worker can
// report an album's last task done right after that task left pendingTasks,
// i.e. the moment dispatch() has nothing left to hand out.
//
// If this test ever fails by timing out, its own t.Cleanup below
// (close(q.quit); <-q.done) will hang too, since a genuinely deadlocked
// dispatch() goroutine can never reach that select - nothing can force it to
// unwind from outside. That's an inherent limit of testing a real deadlock
// in-process: go test's own overall -timeout is what ultimately kills the
// binary in that case, not this test.
func TestCompleteMediaLastTaskDoesNotDeadlockDispatch(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	album := &models.Album{Title: "deadlock repro", Path: t.TempDir()}
	if err := db.Create(album).Error; err != nil {
		t.Fatalf("create album: %v", err)
	}

	q := newQueue(context.Background(), db, 1)

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

	q.medias <- tsk

	select {
	case <-taskDone:
	case <-time.After(5 * time.Second):
		t.Fatal("processMedia never returned for the album's last task - dispatch() likely deadlocked while the worker waited to report the album done")
	}
}

// TestDispatchKeepsAcceptingWorkWhileExpanding checks that a slow album
// expansion doesn't stop dispatch() from receiving new submissions.
func TestDispatchKeepsAcceptingWorkWhileExpanding(t *testing.T) {
	release := make(chan struct{})

	q := newQueue(context.Background(), nil, 1)
	q.expandFn = func(context.Context, *gorm.DB, AlbumRequest, func(int)) ([]*task, *albumState, error) {
		<-release
		return nil, nil, errors.New("expansion failed")
	}

	go q.dispatch()
	t.Cleanup(func() {
		close(release)
		close(q.quit)
		<-q.done
	})

	q.albums <- AlbumRequest{Album: &models.Album{Model: models.Model{ID: 1}}}

	select {
	case q.albums <- AlbumRequest{Album: &models.Album{Model: models.Model{ID: 2}}}:
	case <-time.After(5 * time.Second):
		t.Error("dispatch() stopped accepting submissions while an album was expanding")
	}
}

// TestWorkerIdleTimeoutDoesNotStrandTasks checks that a task submitted after
// the pool's only worker exited on its idle timeout still gets processed.
func TestWorkerIdleTimeoutDoesNotStrandTasks(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	album := &models.Album{Title: "idle worker", Path: t.TempDir()}
	if err := db.Create(album).Error; err != nil {
		t.Fatalf("create album: %v", err)
	}

	q := newQueue(context.Background(), db, 1)
	q.idleTimeout = time.Millisecond
	go q.dispatch()
	t.Cleanup(func() {
		close(q.quit)
		<-q.done
	})

	cache := scanner_cache.MakeAlbumCache()
	for i := range 3 {
		state := newAlbumState(db, album, cache, 1)
		state.skipAfter = true

		// The path doesn't need to exist: processMedia reports the task done
		// however it ended.
		done := make(chan struct{})
		q.medias <- newTask(album, cache, state, filepath.Join(album.Path, "does-not-exist.jpg"), done)

		select {
		case <-done:
		case <-time.After(30 * time.Second):
			t.Fatalf("task %d was never processed", i)
		}

		time.Sleep(50 * time.Millisecond)
	}
}

// TestEnqueueAlbumDedupsUntilDone checks that an album is rejected from when
// it's first accepted until its scan is reported done.
func TestEnqueueAlbumDedupsUntilDone(t *testing.T) {
	q := newQueue(context.Background(), nil, 1)
	req := AlbumRequest{Album: &models.Album{Model: models.Model{ID: 1}}}

	q.enqueueAlbum(req)
	q.enqueueAlbum(req)
	if len(q.pendingAlbums) != 1 {
		t.Fatalf("pendingAlbums = %d, want 1 after submitting the same album twice", len(q.pendingAlbums))
	}

	// Still in-flight once dequeued for expansion.
	q.pendingAlbums = nil
	q.enqueueAlbum(req)
	if len(q.pendingAlbums) != 0 {
		t.Fatalf("pendingAlbums = %d, want 0 while the album is in-flight", len(q.pendingAlbums))
	}

	delete(q.inFlight, req.Album.ID)
	q.enqueueAlbum(req)
	if len(q.pendingAlbums) != 1 {
		t.Fatalf("pendingAlbums = %d, want 1 after the album finished", len(q.pendingAlbums))
	}
}

// TestUrgentTasksDispatchFirst checks that urgent tasks are handed out before
// pending ones, whatever order they arrived in.
func TestUrgentTasksDispatchFirst(t *testing.T) {
	q := newQueue(context.Background(), nil, 1)

	pending, urgent := &task{path: "pending"}, &task{path: "urgent"}
	q.pendingTasks = []*task{pending}
	q.urgentTasks = []*task{urgent}

	for _, want := range []*task{urgent, pending, nil} {
		got := q.nextTask()
		if got != want {
			t.Fatalf("nextTask() = %v, want %v", got, want)
		}
		if got != nil {
			q.popTask()
		}
	}
}
