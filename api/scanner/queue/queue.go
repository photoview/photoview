package queue

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/log"
	"github.com/photoview/photoview/api/scanner"
	"github.com/photoview/photoview/api/scanner/scanner_cache"
	"gorm.io/gorm"
)

type Queue struct {
	db      *gorm.DB
	wait    sync.WaitGroup
	ctx     context.Context
	input   chan []Job
	handout chan Job
	done    chan struct{}
	trigger *time.Ticker

	workers   []*worker
	workersMu sync.Mutex

	backlog []Job // Must be handled by `run` goroutine only.
}

func NewQueue(db *gorm.DB) (*Queue, error) {
	var siteInfo models.SiteInfo
	if err := db.First(&siteInfo).Error; err != nil {
		return nil, fmt.Errorf("can't get site info: %w", err)
	}

	if siteInfo.PeriodicScanInterval < 0 {
		return nil, fmt.Errorf("invalid periodic scan interval (%d): must >=0", siteInfo.PeriodicScanInterval)
	}

	interval := time.Duration(siteInfo.PeriodicScanInterval) * time.Second
	var ticker *time.Ticker
	if interval == 0 {
		ticker = time.NewTicker(time.Second) // The interval is not matter since the ticker is stopped.
		ticker.Stop()
	} else {
		ticker = time.NewTicker(interval)
	}

	if siteInfo.ConcurrentWorkers < 0 {
		return nil, fmt.Errorf("invalid concurrent workers (%d): must >=0", siteInfo.ConcurrentWorkers)
	}

	ret := &Queue{
		db:      db,
		ctx:     log.WithAttrs(context.Background(), "process", "queue"),
		input:   make(chan []Job),
		handout: make(chan Job),
		done:    make(chan struct{}),
		trigger: ticker,
	}

	if err := ret.RescaleWorkers(siteInfo.ConcurrentWorkers); err != nil {
		return nil, err
	}

	ret.wait.Add(1)
	go ret.run()

	return ret, nil
}

func (q *Queue) Close() {
	defer q.wait.Wait()

	q.trigger.Stop()
	close(q.done)
}

func (q *Queue) UpdateScanInterval(newInterval time.Duration) error {
	if newInterval < 0 {
		return fmt.Errorf("invalid periodic scan interval(%v): must >=0", newInterval)
	}

	if newInterval == 0 {
		q.trigger.Stop()
		return nil
	}

	q.trigger.Reset(newInterval)
	return nil
}

func (q *Queue) RescaleWorkers(newMax int) error {
	if newMax < 0 {
		return fmt.Errorf("invalid concurrent workers (%d): must >=0", newMax)
	}

	q.workersMu.Lock()
	defer q.workersMu.Unlock()

	if len(q.workers) == newMax {
		return nil
	}

	if len(q.workers) > newMax {
		closing := q.workers[newMax:]
		q.workers = q.workers[:newMax]

		for _, worker := range closing {
			worker.Close()
		}

		return nil
	}

	// len(q.workers) < newMax
	q.workers = slices.Grow(q.workers, newMax-len(q.workers))
	for len(q.workers) < newMax {
		worker := newWorker(log.WithAttrs(q.ctx, "worker_id", len(q.workers)), q.db, q.handout, &q.wait)
		q.wait.Add(1)
		go worker.Run()
		q.workers = append(q.workers, worker)
	}

	return nil
}

func (q *Queue) run() {
	defer q.RescaleWorkers(0)
	defer q.wait.Done()

MAIN:
	for {
		if len(q.backlog) > 0 {
			select {
			case q.handout <- q.backlog[0]:
				q.backlog = q.backlog[1:]

			case input := <-q.input:
				q.backlog = append(q.backlog, input...)
			case <-q.trigger.C:
				q.fillAllAlbumsToBacklog()
			case <-q.done:
				break MAIN
			}

			continue
		}

		select {
		case input := <-q.input:
			q.backlog = append(q.backlog, input...)
		case <-q.trigger.C:
			q.fillAllAlbumsToBacklog()
		case <-q.done:
			break MAIN
		}
	}
}

func (q *Queue) AddAllAlbums(ctx context.Context) error {
	jobs, err := q.findAllAlbumsJobs()
	if err != nil {
		return err
	}

	select {
	case q.input <- jobs:
	case <-ctx.Done():
		return ctx.Err()
	case <-q.done:
		return fmt.Errorf("queue is closed")
	}

	return nil
}

func (q *Queue) AddUserAlbums(ctx context.Context, user *models.User) error {
	jobs, err := q.findUserAlbumsJobs(user)
	if err != nil {
		return fmt.Errorf("find albums for user (id: %d) error: %w", user.ID, err)
	}

	select {
	case q.input <- jobs:
	case <-ctx.Done():
		return ctx.Err()
	case <-q.done:
		return fmt.Errorf("queue is closed")
	}

	return nil
}

// Must be run in the `run()` goroutine.
func (q *Queue) fillAllAlbumsToBacklog() {
	jobs, err := q.findAllAlbumsJobs()
	if err != nil {
		log.Error(q.ctx, "interval scan", "error", err)
		return
	}

	q.backlog = append(q.backlog, jobs...)
}

func (q *Queue) findAllAlbumsJobs() ([]Job, error) {
	var users []*models.User
	if err := q.db.Find(&users).Error; err != nil {
		return nil, fmt.Errorf("get all users from database error: %w", err)
	}

	var jobs []Job

	for _, user := range users {
		job, err := q.findUserAlbumsJobs(user)
		if err != nil {
			return nil, fmt.Errorf("failed to add user (id: %d) for scanning: %w", user.ID, err)
		}
		jobs = append(jobs, job...)
	}

	return jobs, nil
}

func (q *Queue) findUserAlbumsJobs(user *models.User) ([]Job, error) {
	albumCache := scanner_cache.MakeAlbumCache()
	albums, album_errors := scanner.FindAlbumsForUser(q.db, user, albumCache)
	if len(album_errors) != 0 {
		return nil, Errors(album_errors)
	}

	jobs := make([]Job, 0, len(albums))
	for _, album := range albums {
		jobs = append(jobs, Job{
			album: album,
			cache: albumCache,
		})
	}

	return jobs, nil
}
