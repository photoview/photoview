// Package queue implements the scanner's concurrency and per-media
// processing pipeline. It replaces api/scanner/scanner_queue (the goroutine-
// per-job dispatcher) and the api/scanner/scanner_task + api/scanner/
// scanner_tasks plugin architecture (11 hooks spread across per-purpose
// tasks) with a single dispatcher goroutine driving a self-scaling pool of
// workers, each running every candidate media file through task's four
// phases: gather, decide, process, persist.
package queue

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/scanner_cache"
	"github.com/photoview/photoview/api/scanner/scanner_utils"
	"gorm.io/gorm"
)

// defaultWorkerIdleTimeout is how long a worker waits for its next task
// before exiting, releasing its exiftool process.
const defaultWorkerIdleTimeout = 30 * time.Second

// Queue is a single dispatcher goroutine that turns AlbumRequests into Tasks
// and feeds them to a pool of workers. dispatch() is an event loop that never
// blocks outside its one select: anything slow runs in its own goroutine and
// reports back over a channel. That makes every send to dispatch() safe, and
// means the scheduling state below is only ever touched by dispatch(), so
// none of it needs locking.
type Queue struct {
	ctx context.Context

	db          *gorm.DB
	max         atomic.Int64
	idleTimeout time.Duration
	expandFn    func(ctx context.Context, db *gorm.DB, req AlbumRequest, markDone func(albumID int)) ([]*task, *albumState, error)

	albums chan AlbumRequest
	medias chan *task
	quit   chan struct{}
	done   chan struct{}

	// Reports back to dispatch() from the goroutines it started.
	expanded     chan expandResult
	albumDone    chan int
	workerExited chan struct{}

	// taskChan hands tasks to workers. It's unbuffered, so a worker exiting
	// on its idle timeout can never strand a task, and it's only closed once
	// the queue is shutting down.
	taskChan chan *task

	pendingAlbums []AlbumRequest
	// urgentTasks are dispatched before pendingTasks: a caller is blocked on
	// each of them (see SubmitMedia), while nothing waits on an album's tasks.
	urgentTasks  []*task
	pendingTasks []*task

	// inFlight holds the ID of every album from when it's accepted into
	// pendingAlbums until its scan has finished (see onExpanded and
	// markAlbumDone). Without this, a resubmission of an album mid-scan would
	// expand it a second time concurrently with the first, and whichever
	// albumState's cleanupStaleMedia runs first would delete media the other,
	// still-running scan already found and persisted.
	inFlight map[int]struct{}

	expanding     bool
	activeWorkers int
}

// expandResult is how an expansion goroutine reports back to dispatch().
// tasks is empty if the album had nothing to do or failed to expand; either
// way expand() has already finished with the album.
type expandResult struct {
	albumID int
	tasks   []*task
}

func newQueue(ctx context.Context, db *gorm.DB, workers int) *Queue {
	q := &Queue{
		ctx:          context.WithoutCancel(ctx),
		db:           db,
		idleTimeout:  defaultWorkerIdleTimeout,
		expandFn:     expandAlbum,
		albums:       make(chan AlbumRequest),
		medias:       make(chan *task),
		quit:         make(chan struct{}),
		done:         make(chan struct{}),
		expanded:     make(chan expandResult),
		albumDone:    make(chan int),
		workerExited: make(chan struct{}),
		taskChan:     make(chan *task),
		inFlight:     make(map[int]struct{}),
	}
	q.max.Store(int64(workers))

	return q
}

// markAlbumDone records that albumID's last task has finished (including its
// stale-media cleanup) - called from a worker goroutine via CompleteMedia.
// dispatch() outlives every worker, so this send always gets received.
func (q *Queue) markAlbumDone(albumID int) {
	q.albumDone <- albumID
}

var globalQueue *Queue

// Initialize starts the scanner queue's dispatcher goroutine. ctx is
// typically the application's root context (from main); it is only ever
// used for its values (e.g. logging attributes) - its cancellation is
// stripped, since a canceled ctx must not interrupt in-flight workers.
func Initialize(ctx context.Context, db *gorm.DB) error {
	siteInfo, err := models.GetSiteInfo(db)
	if err != nil {
		return fmt.Errorf("get site info: %w", err)
	}

	q := newQueue(ctx, db, siteInfo.ConcurrentWorkers)

	globalQueue = q
	go q.dispatch()

	return nil
}

// Close signals the dispatcher to stop accepting new work and blocks until
// everything already queued (including in-flight tasks) has finished and
// every worker has exited.
func Close() {
	if globalQueue == nil {
		return
	}

	q := globalQueue
	globalQueue = nil
	close(q.quit)
	<-q.done
}

// ChangeConcurrentWorkers updates the worker pool size. It never forcibly
// shrinks a pool already running; it only changes how far the pool is
// allowed to grow from here on.
func ChangeConcurrentWorkers(n int) {
	if globalQueue == nil {
		return
	}
	globalQueue.max.Store(int64(n))
}

// Initialized reports whether the queue's dispatcher has been started via
// Initialize and not yet stopped by Close.
func Initialized() bool {
	return globalQueue != nil
}

// SubmitAlbum queues a single album for scanning; it is the queue's actual
// unit of work - deciding *which* albums to scan is the caller's job (see
// scanner.AddUser / scanner.AddAll).
func SubmitAlbum(album *models.Album, cache *scanner_cache.AlbumScannerCache) error {
	if globalQueue == nil {
		return fmt.Errorf("scanner queue not initialized")
	}

	select {
	case globalQueue.albums <- AlbumRequest{Album: album, Cache: cache}:
	case <-globalQueue.quit:
		return fmt.Errorf("scanner queue is shutting down")
	}

	return nil
}

// SubmitMedia (re)processes a single already-known media, submitting it to
// the same worker pool as everything else, but ahead of any queued album
// tasks - used to fix up one file (e.g. regenerate a cache file found missing
// while serving it) without starting a dedicated worker for it. It blocks
// until processing has finished.
func SubmitMedia(ctx context.Context, db *gorm.DB, album *models.Album, cache *scanner_cache.AlbumScannerCache, mediaPath string) error {
	if globalQueue == nil {
		return fmt.Errorf("scanner queue not initialized")
	}

	state := newAlbumState(db, album, cache, 1)
	state.skipAfter = true

	t := newTask(album, cache, state, mediaPath, make(chan struct{}))

	select {
	case globalQueue.medias <- t:
	case <-ctx.Done():
		return ctx.Err()
	case <-globalQueue.quit:
		return fmt.Errorf("scanner queue is shutting down")
	}

	select {
	case <-t.done:
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

func (q *Queue) dispatch() {
	defer close(q.done)

	albums, medias, quit := q.albums, q.medias, q.quit
	closing, workersStopped := false, false

	for {
		q.startExpansion()

		next := q.nextTask()
		if next != nil {
			q.ensureWorker()
		}

		if closing && next == nil && len(q.pendingAlbums) == 0 && !q.expanding {
			if !workersStopped {
				close(q.taskChan)
				workersStopped = true
			}
			if q.activeWorkers == 0 {
				return
			}
		}

		// A nil channel is never ready, so with no task to send this case
		// just never fires.
		var out chan<- *task
		if next != nil {
			out = q.taskChan
		}

		select {
		case out <- next:
			q.popTask()
		case req := <-albums:
			q.enqueueAlbum(req)
		case t := <-medias:
			q.urgentTasks = append(q.urgentTasks, t)
		case res := <-q.expanded:
			q.onExpanded(res)
		case id := <-q.albumDone:
			delete(q.inFlight, id)
		case <-q.workerExited:
			q.activeWorkers--
		case <-quit:
			// Submitters select on quit too, so none of them get stuck once
			// these stop being received from. This is also what guarantees
			// nothing new can need a worker after taskChan is closed.
			closing = true
			albums, medias, quit = nil, nil, nil
		}
	}
}

// startExpansion starts expanding the next pending album, if nothing is
// expanding already. Pending tasks take priority over expanding more albums.
func (q *Queue) startExpansion() {
	if q.expanding || len(q.pendingTasks) > 0 || len(q.pendingAlbums) == 0 {
		return
	}

	req := q.pendingAlbums[0]
	q.pendingAlbums = q.pendingAlbums[1:]
	q.expanding = true

	go q.expand(req)
}

// expand runs in its own goroutine since it can be slow: listing the album
// and identifying each file's media type, and for an album with nothing to
// do, its stale-media cleanup too.
func (q *Queue) expand(req AlbumRequest) {
	tasks, state, err := q.expandFn(q.ctx, q.db, req, q.markAlbumDone)
	switch {
	case err != nil:
		scanner_utils.ScannerError(q.ctx, "expand album (%s): %s", req.Album.Path, err)
	case len(tasks) == 0:
		state.completeAlbum(q.ctx)
	}

	q.expanded <- expandResult{albumID: req.Album.ID, tasks: tasks}
}

func (q *Queue) onExpanded(res expandResult) {
	q.expanding = false

	if len(res.tasks) == 0 {
		// Nothing will ever complete this album's (nonexistent) tasks, so
		// nothing else will report it done - without this, the album could
		// never be resubmitted.
		delete(q.inFlight, res.albumID)
		return
	}

	q.pendingTasks = append(q.pendingTasks, res.tasks...)
}

func (q *Queue) enqueueAlbum(req AlbumRequest) {
	if _, ok := q.inFlight[req.Album.ID]; ok {
		return
	}

	q.inFlight[req.Album.ID] = struct{}{}
	q.pendingAlbums = append(q.pendingAlbums, req)
}

// nextTask returns the task to dispatch next, or nil if there is none.
func (q *Queue) nextTask() *task {
	if len(q.urgentTasks) > 0 {
		return q.urgentTasks[0]
	}
	if len(q.pendingTasks) > 0 {
		return q.pendingTasks[0]
	}
	return nil
}

// popTask removes the task nextTask returned.
func (q *Queue) popTask() {
	if len(q.urgentTasks) > 0 {
		q.urgentTasks = q.urgentTasks[1:]
		return
	}
	q.pendingTasks = q.pendingTasks[1:]
}

// ensureWorker grows the pool by one worker, up to max. Calling this once per
// dispatch() iteration is enough: a task only ever waits on a pool that's
// either already full or has just gained a worker to take it.
func (q *Queue) ensureWorker() {
	if q.activeWorkers >= int(q.max.Load()) {
		return
	}

	q.activeWorkers++
	go q.runWorker()
}

// runWorker processes tasks until taskChan is closed or no task shows up for
// idleTimeout. Either way, its exit is reported to dispatch() only after the
// worker itself has shut down, so that Close() really does wait for it.
func (q *Queue) runWorker() {
	defer func() { q.workerExited <- struct{}{} }()

	w := newWorker(q.ctx, q.db)
	defer w.close()

	idle := time.NewTimer(q.idleTimeout)
	defer idle.Stop()

	for {
		select {
		case t, ok := <-q.taskChan:
			if !ok {
				return
			}
			w.processMedia(t)
			idle.Reset(q.idleTimeout)
		case <-idle.C:
			return
		}
	}
}
