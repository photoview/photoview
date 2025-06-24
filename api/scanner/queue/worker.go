package queue

import (
	"context"

	"github.com/photoview/photoview/api/log"
)

type workerCallback[Job any] interface {
	processJob(ctx context.Context, job Job)
	finish(ctx context.Context)
}

// The worker processes all jobs in the same goroutine.
type worker[Job any] struct {
	ctx      context.Context
	input    <-chan Job
	done     chan struct{}
	callback workerCallback[Job]
}

func newWorker[Job any](ctx context.Context, input <-chan Job, callback workerCallback[Job]) *worker[Job] {
	return &worker[Job]{
		ctx:      ctx,
		input:    input,
		done:     make(chan struct{}),
		callback: callback,
	}
}

func (w *worker[Job]) Close() {
	log.Info(w.ctx, "closing worker")
	close(w.done)
}

func (w *worker[Job]) Run() {
	defer w.callback.finish(w.ctx)
	defer log.Info(w.ctx, "worker done")

	log.Info(w.ctx, "worker start")

MAIN:
	for {
		select {
		case job := <-w.input:
			log.Info(w.ctx, "handout", "job", job)
			w.callback.processJob(w.ctx, job)
		case <-w.done:
			break MAIN
		}
	}
}
