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
	"sync"
	"sync/atomic"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/scanner_cache"
	"github.com/photoview/photoview/api/scanner/scanner_utils"
	"gorm.io/gorm"
)

// Queue is a single dispatcher goroutine that turns AlbumRequests into Tasks
// and feeds them to a pool of workers. pendingAlbums, pendingTasks, taskChan,
// activeWorkers and scanningAlbums are only ever touched by dispatch(), so
// none of them need locking of their own. finishedAlbums is the one
// exception - see its own doc comment.
type Queue struct {
	ctx context.Context

	db  *gorm.DB
	max atomic.Int64

	incoming chan any // AlbumRequest or *task
	quit     chan struct{}
	done     chan struct{}

	pendingAlbums []AlbumRequest
	pendingTasks  []*task

	// scanningAlbums holds the ID of every album that has been dequeued from
	// pendingAlbums for expansion but hasn't yet finished (expandAlbum error,
	// zero tasks, or its albumState's last task completing - see the
	// dispatch() branches and markAlbumDone). Without this, enqueue()'s
	// AlbumRequest dedup only sees pendingAlbums, which is empty for exactly
	// this album the moment it's mid-scan - a resubmission arriving in that
	// window would expand the same album a second time concurrently with the
	// first, and whichever albumState's cleanupStaleMedia runs first would
	// delete media the other, still-running scan already found and persisted.
	scanningAlbums map[int]struct{}

	// finishedAlbums/finishedMu are how a worker goroutine reports an
	// album's last task completing (see markAlbumDone) - a mutex-guarded
	// append rather than a send on incoming, because dispatch() can
	// legitimately be parked in wg.Wait() below (tearing down idle workers)
	// at exactly the moment a worker needs to report in, and a channel send
	// would deadlock the two goroutines waiting on each other. enqueue()
	// drains this into scanningAlbums before every dedup check.
	finishedMu     sync.Mutex
	finishedAlbums []int

	// taskChan is created lazily and closed (then recreated on next use)
	// every time the pool goes idle - that close is what tells its workers
	// to exit their `range` loop and return.
	taskChan      chan *task
	activeWorkers int
	wg            sync.WaitGroup // whole-Queue lifetime; Add/Wait only ever called from dispatch()
}

// markAlbumDone records that albumID's last task has finished (including its
// stale-media cleanup) - called from a worker goroutine via CompleteMedia,
// so it must never block on dispatch() being in any particular state.
func (q *Queue) markAlbumDone(albumID int) {
	q.finishedMu.Lock()
	q.finishedAlbums = append(q.finishedAlbums, albumID)
	q.finishedMu.Unlock()
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

	q := &Queue{
		ctx:            context.WithoutCancel(ctx),
		db:             db,
		incoming:       make(chan any),
		quit:           make(chan struct{}),
		done:           make(chan struct{}),
		scanningAlbums: make(map[int]struct{}),
	}
	q.max.Store(int64(siteInfo.ConcurrentWorkers))

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
	case globalQueue.incoming <- AlbumRequest{Album: album, Cache: cache}:
	case <-globalQueue.quit:
		return fmt.Errorf("scanner queue is shutting down")
	}

	return nil
}

// SubmitMedia (re)processes a single already-known media, submitting it to
// the same queue and worker pool as everything else - used to fix up one
// file (e.g. regenerate a cache file found missing while serving it) without
// starting a dedicated worker for it. It blocks until processing has
// finished.
func SubmitMedia(ctx context.Context, db *gorm.DB, album *models.Album, cache *scanner_cache.AlbumScannerCache, mediaPath string) error {
	if globalQueue == nil {
		return fmt.Errorf("scanner queue not initialized")
	}

	state := newAlbumState(db, album, cache, 1)
	state.skipAfter = true

	t := newTask(album, cache, state, mediaPath, make(chan struct{}))

	select {
	case globalQueue.incoming <- t:
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

	for {
		// Pending tasks take priority over expanding more albums.
		if len(q.pendingTasks) == 0 && len(q.pendingAlbums) > 0 {
			req := q.pendingAlbums[0]
			q.pendingAlbums = q.pendingAlbums[1:]
			q.scanningAlbums[req.Album.ID] = struct{}{}

			tasks, state, err := expandAlbum(q.ctx, q.db, req, q.markAlbumDone)
			switch {
			case err != nil:
				scanner_utils.ScannerError(q.ctx, "expand album (%s): %s", req.Album.Path, err)
				// Nothing will ever complete this album's (nonexistent)
				// tasks, so nothing will report it done - without this, the
				// album could never be resubmitted.
				delete(q.scanningAlbums, req.Album.ID)
			case len(tasks) == 0:
				// Nothing to do. This runs synchronously, in dispatch()
				// itself, so scanningAlbums can be cleared directly.
				state.completeAlbum(q.ctx)
				delete(q.scanningAlbums, req.Album.ID)
			default:
				q.pendingTasks = append(q.pendingTasks, tasks...)
			}
			continue
		}

		if len(q.pendingTasks) == 0 {
			if q.taskChan != nil {
				// No work left, close all workers
				close(q.taskChan)
				q.taskChan, q.activeWorkers = nil, 0
				q.wg.Wait()
			}

			// Both pendingAlbums and pendingTasks are empty.
			select {
			case req := <-q.incoming:
				q.enqueue(req)
			case <-q.quit:
				return
			}
			continue
		}

		// Ensure workers are running
		if q.taskChan == nil {
			q.taskChan = make(chan *task)
		}
		if q.activeWorkers < int(q.max.Load()) {
			q.activeWorkers++
			q.wg.Add(1)
			go func(ch <-chan *task) {
				defer q.wg.Done()
				w := newWorker(q.ctx, q.db)
				defer w.close()
				for t := range ch {
					w.processMedia(t)
				}
			}(q.taskChan)
		}

		select {
		case q.taskChan <- q.pendingTasks[0]:
			q.pendingTasks = q.pendingTasks[1:]
		case req := <-q.incoming:
			q.enqueue(req)
		}
	}
}

func (q *Queue) enqueue(req any) {
	// Drain any albums a worker has reported done since we last checked,
	// before the AlbumRequest dedup below reads scanningAlbums - otherwise a
	// resubmission could be rejected as "still scanning" long after that
	// scan actually finished, with no retry to recover it.
	q.finishedMu.Lock()
	for _, id := range q.finishedAlbums {
		delete(q.scanningAlbums, id)
	}
	q.finishedAlbums = q.finishedAlbums[:0]
	q.finishedMu.Unlock()

	switch r := req.(type) {
	case AlbumRequest:
		if _, scanning := q.scanningAlbums[r.Album.ID]; scanning {
			return
		}
		for _, pending := range q.pendingAlbums {
			if pending.Album.ID == r.Album.ID {
				return
			}
		}
		q.pendingAlbums = append(q.pendingAlbums, r)
	case *task:
		q.pendingTasks = append(q.pendingTasks, r)
	}
}
