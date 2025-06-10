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
	db          *gorm.DB
	waitWorkers sync.WaitGroup
	ctx         context.Context
	input       chan []Job
	handout     chan Job
	done        chan struct{}
	trigger     *time.Ticker
	backlog     []Job

	workers   []*worker
	workersMu sync.Mutex
}

func NewQueue(db *gorm.DB) (*Queue, error) {
	var siteInfo models.SiteInfo
	if err := db.First(&siteInfo).Error; err != nil {
		return nil, fmt.Errorf("can't get site info: %w", err)
	}

	interval := time.Duration(siteInfo.PeriodicScanInterval) * time.Second

	ret := &Queue{
		db:      db,
		ctx:     log.WithAttrs(context.Background(), "process", "queue"),
		input:   make(chan []Job),
		handout: make(chan Job),
		done:    make(chan struct{}),
		trigger: time.NewTicker(interval),
	}

	ret.UpdateWorkers(siteInfo.ConcurrentWorkers)

	return ret, nil
}

func (q *Queue) Close() {
	q.trigger.Stop()
	close(q.done)
}

func (q *Queue) UpdateInterval(newInterval time.Duration) {
	q.trigger.Reset(newInterval)
}

func (q *Queue) UpdateWorkers(newMax int) {
	if newMax < 0 {
		newMax = 0
	}

	q.workersMu.Lock()
	defer q.workersMu.Unlock()

	if len(q.workers) == newMax {
		return
	}

	if len(q.workers) > newMax {
		closing := q.workers[newMax:]
		q.workers = q.workers[:newMax]

		for _, worker := range closing {
			worker.Close()
		}

		return
	}

	// len(q.workers) < newMax
	q.workers = slices.Grow(q.workers, newMax-len(q.workers))
	for len(q.workers) < newMax {
		worker := newWorker(log.WithAttrs(q.ctx, "worker_id", len(q.workers)), q.db, q.handout, &q.waitWorkers)
		q.waitWorkers.Add(1)
		go worker.Run()
		q.workers = append(q.workers, worker)
	}
}

func (q *Queue) Run() {
	defer func() {
		q.UpdateWorkers(0)
		q.waitWorkers.Wait()
	}()

  MAIN:
	for {
		if len(q.backlog) > 0 {
			select {
			case q.handout <- q.backlog[0]:
				q.backlog = q.backlog[1:]

			case input := <-q.input:
				q.backlog = append(q.backlog, input...)
			case <-q.trigger.C:
				q.AddAllUsers(q.ctx)
			case <-q.done:
				break MAIN
			}

			continue
		}

		select {
		case input := <-q.input:
			q.backlog = append(q.backlog, input...)
		case <-q.trigger.C:
			q.AddAllUsers(q.ctx)
		case <-q.done:
			break MAIN
		}
	}
}

func (q *Queue) AddAllUsers(ctx context.Context) error {
	var users []*models.User
	if err := q.db.Find(&users).Error; err != nil {
		return fmt.Errorf("get all users from database error: %w", err)
	}

	var jobs []Job

	for _, user := range users {
		job, err := q.findUserJobs(user)
		if err != nil {
			return fmt.Errorf("failed to add user (id: %d) for scanning: %w", user.ID, err)
		}
		jobs = append(jobs, job...)
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

func (q *Queue) AddUser(ctx context.Context, user *models.User) error {
	jobs, err := q.findUserJobs(user)
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

func (q *Queue) findUserJobs(user *models.User) ([]Job, error) {
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
