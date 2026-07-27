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
	"github.com/photoview/photoview/api/scanner"
	"github.com/photoview/photoview/api/scanner/scanner_cache"
	"github.com/photoview/photoview/api/scanner/scanner_utils"
	"gorm.io/gorm"
)

// Queue is a single dispatcher goroutine that turns AlbumRequests into Tasks
// and feeds them to a pool of workers. pendingAlbums, pendingTasks, taskChan
// and activeWorkers are only ever touched by dispatch(), so none of them need
// locking of their own.
type Queue struct {
	ctx context.Context

	db  *gorm.DB
	max atomic.Int64

	incoming chan any // AlbumRequest or *task
	quit     chan struct{}
	done     chan struct{}

	pendingAlbums []AlbumRequest
	pendingTasks  []*task

	// taskChan is created lazily and closed (then recreated on next use)
	// every time the pool goes idle - that close is what tells its workers
	// to exit their `range` loop and return.
	taskChan      chan *task
	activeWorkers int
	wg            sync.WaitGroup // whole-Queue lifetime; Add/Wait only ever called from dispatch()
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
		ctx:      context.WithoutCancel(ctx),
		db:       db,
		incoming: make(chan any),
		quit:     make(chan struct{}),
		done:     make(chan struct{}),
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
	close(globalQueue.quit)
	<-globalQueue.done
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

// AddUser finds all root albums owned by user and queues them for scanning.
// It does not block until scanning finishes.
func AddUser(user *models.User) error {
	if globalQueue == nil {
		return fmt.Errorf("scanner queue not initialized")
	}

	cache := scanner_cache.MakeAlbumCache()
	albums, errs := scanner.FindAlbumsForUser(globalQueue.db, user, cache)
	for _, err := range errs {
		if err != nil {
			return fmt.Errorf("find albums for user (user_id: %d): %w", user.ID, err)
		}
	}

	for _, album := range albums {
		select {
		case globalQueue.incoming <- AlbumRequest{Album: album, Cache: cache}:
		case <-globalQueue.quit:
			return fmt.Errorf("scanner queue is shutting down")
		}
	}

	return nil
}

// AddAll queues every user's albums for scanning.
func AddAll() error {
	if globalQueue == nil {
		return fmt.Errorf("scanner queue not initialized")
	}

	var users []*models.User
	if err := globalQueue.db.Find(&users).Error; err != nil {
		return fmt.Errorf("get all users from database: %w", err)
	}

	for _, user := range users {
		if err := AddUser(user); err != nil {
			return fmt.Errorf("add user to queue (user_id: %d): %w", user.ID, err)
		}
	}

	return nil
}

// ProcessMedia (re)processes a single already-known media, submitting it to
// the same queue and worker pool as everything else - used to fix up one
// file (e.g. regenerate a cache file found missing while serving it) without
// starting a dedicated worker for it. It blocks until processing has
// finished.
func ProcessMedia(ctx context.Context, db *gorm.DB, media *models.Media) error {
	if globalQueue == nil {
		return fmt.Errorf("scanner queue not initialized")
	}

	var album models.Album
	if err := db.Model(media).Association("Album").Find(&album); err != nil {
		return err
	}

	cache := scanner_cache.MakeAlbumCache()
	state := newAlbumState(db, &album, cache, 1)
	state.skipAfter = true

	t := newTask(&album, cache, state, media.Path, make(chan struct{}))

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

			tasks, state, err := expandAlbum(q.ctx, q.db, req)
			switch {
			case err != nil:
				scanner_utils.ScannerError(q.ctx, "expand album (%s): %s", req.Album.Path, err)
			case len(tasks) == 0:
				// Nothing to do
				state.completeAlbum(q.ctx)
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
	switch r := req.(type) {
	case AlbumRequest:
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
