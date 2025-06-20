package queue

import (
	"context"
	"sync"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/log"
	"github.com/photoview/photoview/api/scanner"
	"github.com/photoview/photoview/api/scanner/scanner_cache"
	"github.com/photoview/photoview/api/scanner/scanner_task"
	"github.com/photoview/photoview/api/scanner/scanner_utils"
	"gorm.io/gorm"
)

type Job struct {
	album *models.Album
	cache *scanner_cache.AlbumScannerCache
}

// The worker processes all jobs in the same goroutine.
type worker struct {
	ctx          context.Context
	db           *gorm.DB
	input        <-chan Job
	done         chan struct{}
	doneCallback func(Job)
	parentWaiter *sync.WaitGroup
}

func newWorker(ctx context.Context, db *gorm.DB, input <-chan Job, callback func(Job), parentWaiter *sync.WaitGroup) *worker {
	return &worker{
		ctx:          ctx,
		db:           db,
		input:        input,
		done:         make(chan struct{}),
		doneCallback: callback,
		parentWaiter: parentWaiter,
	}
}

func (w *worker) Close() {
	log.Info(w.ctx, "closing worker")
	close(w.done)
}

func (w *worker) Run() {
	defer w.parentWaiter.Done()
	defer log.Info(w.ctx, "worker done")

	log.Info(w.ctx, "worker start")

MAIN:
	for {
		select {
		case job := <-w.input:
			w.processJob(job)
		case <-w.done:
			break MAIN
		}
	}
}

func (w *worker) processJob(job Job) {
	log.Info(w.ctx, "process album", "album", job.album.Title)
	defer w.doneCallback(job)

	task := scanner_task.NewTaskContext(w.ctx, w.db, job.album, job.cache)
	if err := scanner.ScanAlbum(task); err != nil {
		scanner_utils.ScannerError(w.ctx, "Failed to scan album: %v", err)
	}
}
