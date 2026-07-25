package queue

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/face_detection"
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

	if err := AddUser(&models.User{}); err == nil {
		t.Errorf("AddUser() error = nil, want an error when the queue is not initialized")
	}

	if err := AddAll(); err == nil {
		t.Errorf("AddAll() error = nil, want an error when the queue is not initialized")
	}

	if err := ProcessMedia(context.Background(), nil, &models.Media{}); err == nil {
		t.Errorf("ProcessMedia() error = nil, want an error when the queue is not initialized")
	}
}

// TestAddAllQueuesEveryUser checks that AddAll finds and scans every user's
// albums, not just one - the loop in AddAll is otherwise indistinguishable
// from AddUser in test coverage.
func TestAddAllQueuesEveryUser(t *testing.T) {
	test_utils.FilesystemTest(t)
	db := test_utils.DatabaseTest(t)

	pass := "1234"
	user1, err := models.RegisterUser(db, "add_all_user_1", &pass, true)
	if err != nil {
		t.Fatalf("register user1: %v", err)
	}
	user2, err := models.RegisterUser(db, "add_all_user_2", &pass, true)
	if err != nil {
		t.Fatalf("register user2: %v", err)
	}

	album1Dir := t.TempDir()
	copyFixtureFile(t, "photo/plain.jpg", filepath.Join(album1Dir, "plain.jpg"))
	album1 := &models.Album{Title: "add-all album 1", Path: album1Dir}
	if err := db.Save(album1).Error; err != nil {
		t.Fatalf("save album1: %v", err)
	}
	if err := db.Model(user1).Association("Albums").Append(album1); err != nil {
		t.Fatalf("associate album1 with user1: %v", err)
	}

	album2Dir := t.TempDir()
	copyFixtureFile(t, "photo/plain.jpg", filepath.Join(album2Dir, "plain.jpg"))
	album2 := &models.Album{Title: "add-all album 2", Path: album2Dir}
	if err := db.Save(album2).Error; err != nil {
		t.Fatalf("save album2: %v", err)
	}
	if err := db.Model(user2).Association("Albums").Append(album2); err != nil {
		t.Fatalf("associate album2 with user2: %v", err)
	}

	if err := face_detection.InitializeFaceDetector(db); err != nil {
		t.Fatalf("initialize face detector: %v", err)
	}

	if err := Initialize(context.Background(), db); err != nil {
		t.Fatalf("Initialize() error: %v", err)
	}
	// Close() unconditionally, even if AddAll() failed, so the dispatcher
	// goroutine it started is never left running past this test.
	addAllErr := AddAll()
	Close()
	if addAllErr != nil {
		t.Fatalf("AddAll() error: %v", addAllErr)
	}

	var count1, count2 int64
	if err := db.Model(&models.Media{}).Where("album_id = ?", album1.ID).Count(&count1).Error; err != nil {
		t.Fatalf("count album1 media: %v", err)
	}
	if err := db.Model(&models.Media{}).Where("album_id = ?", album2.ID).Count(&count2).Error; err != nil {
		t.Fatalf("count album2 media: %v", err)
	}
	if count1 == 0 {
		t.Errorf("expected AddAll to have scanned user1's album, found 0 media")
	}
	if count2 == 0 {
		t.Errorf("expected AddAll to have scanned user2's album, found 0 media")
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
		ctx:      context.Background(),
		db:       db,
		incoming: make(chan any),
		quit:     make(chan struct{}),
		done:     make(chan struct{}),
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
