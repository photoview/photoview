package queue

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/photoview/photoview/api/log"
)

type commonJob interface {
	// Key returns a unique integer of each job, to identify the job.
	Key() int

	// String returns a short description of the job, for logging or any other string output.
	String() string
}

type queueCallback[Job commonJob] interface {
	processJob(ctx context.Context, job Job)
	fillPeriodicJobs(ctx context.Context)
}

type commonQueue[Job commonJob] struct {
	wait     sync.WaitGroup
	ctx      context.Context
	handout  chan Job
	done     chan struct{}
	trigger  *time.Ticker
	callback queueCallback[Job]

	workers   []*worker[Job]
	workersMu sync.Mutex

	backlogUpdated chan struct{}

	backlog []Job
	ongoing map[int]Job
	jobsMu  sync.Mutex
}

func newCommonQueue[Job commonJob](ctx context.Context, interval time.Duration, workerNum int, callback queueCallback[Job]) (*commonQueue[Job], error) {
	if interval < 0 {
		return nil, fmt.Errorf("interval should be >= 0")
	}

	var ticker *time.Ticker
	if interval == 0 {
		ticker = time.NewTicker(time.Second) // The interval does not matter since the ticker is stopped.
		ticker.Stop()
	} else {
		ticker = time.NewTicker(interval)
	}

	ret := &commonQueue[Job]{
		ctx:            ctx,
		handout:        make(chan Job),
		done:           make(chan struct{}),
		trigger:        ticker,
		callback:       callback,
		backlogUpdated: make(chan struct{}, 1),
		ongoing:        make(map[int]Job),
	}

	if err := ret.RescaleWorkers(workerNum); err != nil {
		return nil, err
	}

	return ret, nil
}

// Close closes the queue and waits all workers finishing their jobs. It could be called multiple times in different goroutines.
func (q *commonQueue[Job]) Close() {
	defer func() {
		q.wait.Wait()
		log.Info(q.ctx, "closed queue", "remain_jobs", q.lenJobs())
	}()

	log.Info(q.ctx, "closing queue")
	q.trigger.Stop()
	close(q.done)
	if err := q.RescaleWorkers(0); err != nil {
		log.Error(q.ctx, "failed to close all workers", "error", err)
	}
}

// RunBackground runs the background goroutine to process jobs and periodic jobs.
func (q *commonQueue[Job]) RunBackground() {
	q.wait.Add(1)
	go q.run()
}

// ConsumeAllBacklog waits all jobs to be done in the queue backlog. It doesn't require `RunBackground`.
// This function is useful with unit tests.
func (q *commonQueue[Job]) ConsumeAllBacklog(ctx context.Context) {
	defer log.Info(q.ctx, "empty backlog")

	for {
		job, ok := q.popBacklog()
		if !ok {
			return
		}

		log.Info(q.ctx, "consuming backlog", "job", job)
		select {
		case q.handout <- job:
		case <-ctx.Done():
			q.pushBacklog(job)
			return
		}
	}
}

// UpdateScanInterval updates the interval of background periodic jobs.
func (q *commonQueue[Job]) UpdateScanInterval(newInterval time.Duration) error {
	if newInterval < 0 {
		return fmt.Errorf("invalid periodic scan interval(%d): must be >= 0", newInterval)
	}

	if newInterval == 0 {
		log.Info(q.ctx, "Periodic scan interval changed", "interval", "disabled")
		q.trigger.Stop()
		return nil
	}

	log.Info(q.ctx, "Periodic scan interval changed", "interval", newInterval)
	q.trigger.Reset(newInterval)
	return nil
}

// RescaleWorkers rescales the number of background workers.
func (q *commonQueue[Job]) RescaleWorkers(newMax int) error {
	if newMax < 0 {
		return fmt.Errorf("invalid concurrent workers (%d): must be >= 0", newMax)
	}

	q.workersMu.Lock()
	defer q.workersMu.Unlock()

	defer func() {
		log.Info(q.ctx, "rescaled workers", "workers_number", len(q.workers))
	}()

	if len(q.workers) == newMax {
		return nil
	}

	if len(q.workers) > newMax {
		closing := q.workers[newMax:]
		if newMax == 0 {
			q.workers = nil
		} else {
			q.workers = q.workers[:newMax]
		}

		for _, worker := range closing {
			worker.Close()
		}

		return nil
	}

	// len(q.workers) < newMax
	q.workers = slices.Grow(q.workers, newMax-len(q.workers))
	for len(q.workers) < newMax {
		worker := newWorker(log.WithAttrs(q.ctx, "worker_id", len(q.workers)), q.handout, q)
		q.wait.Add(1)
		go worker.Run()
		q.workers = append(q.workers, worker)
	}

	return nil
}

func (q *commonQueue[Job]) run() {
	defer func() {
		log.Info(q.ctx, "queue background done")
		q.wait.Done()
	}()

	log.Info(q.ctx, "queue background start")
MAIN:
	for {
		log.Info(q.ctx, "backlog length", "len", q.lenJobs())
		job, ok := q.popBacklog()

		if ok {
			handed := false
			done := false
			select {
			case q.handout <- job:
				handed = true
			case <-q.trigger.C:
				q.callback.fillPeriodicJobs(q.ctx)
			case <-q.done:
				done = true
			}

			if !handed {
				// Interrupted by other signal, put the job back.
				// Should not use `appendBacklog()` to avoid signal of `backlogUpdated`.
				q.pushBacklog(job)
			}

			if done {
				break MAIN
			}

			continue
		}

		select {
		case <-q.backlogUpdated:
		case <-q.trigger.C:
			q.callback.fillPeriodicJobs(q.ctx)
		case <-q.done:
			break MAIN
		}
	}
}

func (q *commonQueue[Job]) appendBacklog(jobs []Job) {
	q.jobsMu.Lock()
	defer q.jobsMu.Unlock()

NEXT_NEW_JOB:
	for _, newJob := range jobs {
		if _, ok := q.ongoing[newJob.Key()]; ok {
			continue NEXT_NEW_JOB
		}

		for _, existJob := range q.backlog {
			if existJob.Key() == newJob.Key() {
				continue NEXT_NEW_JOB
			}
		}

		q.backlog = append(q.backlog, newJob)
		log.Info(q.ctx, "insert to queue backlog", "job", newJob)
	}

	select {
	case q.backlogUpdated <- struct{}{}:
	default:
	}
}

func (q *commonQueue[Job]) jobDone(job Job) {
	q.jobsMu.Lock()
	defer q.jobsMu.Unlock()

	delete(q.ongoing, job.Key())
}

func (q *commonQueue[Job]) popBacklog() (Job, bool) {
	q.jobsMu.Lock()
	defer q.jobsMu.Unlock()

	if len(q.backlog) == 0 {
		var ret Job
		return ret, false
	}

	ret := q.backlog[0]
	q.backlog = q.backlog[1:]

	q.ongoing[ret.Key()] = ret

	return ret, true
}

func (q *commonQueue[Job]) pushBacklog(job Job) {
	q.jobsMu.Lock()
	defer q.jobsMu.Unlock()

	delete(q.ongoing, job.Key())
	q.backlog = append([]Job{job}, q.backlog...)
}

func (q *commonQueue[Job]) lenJobs() int {
	q.jobsMu.Lock()
	defer q.jobsMu.Unlock()

	return len(q.backlog) + len(q.ongoing)
}

func (q *commonQueue[Job]) processJob(ctx context.Context, job Job) {
	defer func() {
		if r := recover(); r != nil {
			log.Error(ctx, "panic happened during job processing", "job", job, "panic", r)
		}

		q.jobDone(job)
		log.Info(ctx, "job is done", "job", job)
	}()

	log.Info(ctx, "job is running", "job", job)
	q.callback.processJob(ctx, job)
}

func (q *commonQueue[Job]) finish(ctx context.Context) {
	q.wait.Done()
}
