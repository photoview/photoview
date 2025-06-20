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
	handout chan Job
	done    chan struct{}
	trigger *time.Ticker

	workers   []*worker
	workersMu sync.Mutex

	backlogUpdated chan struct{}

	backlog []Job
	ongoing map[int]Job
	jobsMu  sync.Mutex
}

func NewQueue(db *gorm.DB) (*Queue, error) {
	siteInfo, err := models.GetSiteInfo(db)
	if err != nil {
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
		db:             db,
		ctx:            log.WithAttrs(context.Background(), "process", "queue"),
		handout:        make(chan Job),
		done:           make(chan struct{}),
		trigger:        ticker,
		backlogUpdated: make(chan struct{}, 1),
		ongoing:        make(map[int]Job),
	}

	if err := ret.RescaleWorkers(siteInfo.ConcurrentWorkers); err != nil {
		return nil, err
	}

	return ret, nil
}

func (q *Queue) Close() {
	defer func() {
		q.wait.Wait()
		log.Info(q.ctx, "backlog remain", "length", q.lenBacklog())
	}()

	log.Info(q.ctx, "closing queue")
	q.trigger.Stop()
	close(q.done)
	q.RescaleWorkers(0)
}

func (q *Queue) RunBackground() {
	q.wait.Add(1)
	go q.run()
}

func (q *Queue) ConsumeAllBacklog(ctx context.Context) {
	for {
		job, ok := q.popBacklog()
		if !ok {
			log.Info(q.ctx, "consume all backlog: return")
			return
		}

		log.Info(q.ctx, "consume all backlog", "album", job.album.Title)
		select {
		case q.handout <- job:
		case <-ctx.Done():
			return
		}
	}
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
		worker := newWorker(log.WithAttrs(q.ctx, "worker_id", len(q.workers)), q.db, q.handout, q.jobDone, &q.wait)
		q.wait.Add(1)
		go worker.Run()
		q.workers = append(q.workers, worker)
	}

	return nil
}

func (q *Queue) run() {
	defer q.wait.Done()
	defer log.Info(q.ctx, "queue background done")

	log.Info(q.ctx, "queue background start")
MAIN:
	for {
		job, ok := q.popBacklog()

		if ok {
			log.Info(q.ctx, "run", "album", job.album.Title)
			select {
			case q.handout <- job:
			case <-q.backlogUpdated:
			case <-q.trigger.C:
				q.AddAllAlbums(q.ctx)
			case <-q.done:
				break MAIN
			}

			continue
		}

		select {
		case <-q.backlogUpdated:
		case <-q.trigger.C:
			q.AddAllAlbums(q.ctx)
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

	q.appendBacklog(jobs)

	return nil
}

func (q *Queue) AddUserAlbums(ctx context.Context, user *models.User) error {
	jobs, err := q.findUserAlbumsJobs(user)
	if err != nil {
		return fmt.Errorf("find albums for user (id: %d) error: %w", user.ID, err)
	}

	q.appendBacklog(jobs)

	return nil
}

func (q *Queue) appendBacklog(jobs []Job) {
	q.jobsMu.Lock()
	defer q.jobsMu.Unlock()

NEXT_NEW_JOB:
	for _, newJob := range jobs {
		for _, existJob := range q.ongoing {
			if newJob.album.ID == existJob.album.ID {
				continue NEXT_NEW_JOB
			}
		}

		for _, existJob := range q.backlog {
			if newJob.album.ID == existJob.album.ID {
				continue NEXT_NEW_JOB
			}
		}

		q.backlog = append(q.backlog, newJob)
		log.Info(q.ctx, "insert to queue backlog", "album", newJob.album.Title)
	}

	select {
	case q.backlogUpdated <- struct{}{}:
	default:
	}
}

func (q *Queue) jobDone(job Job) {
	q.jobsMu.Lock()
	defer q.jobsMu.Unlock()

	delete(q.ongoing, job.album.ID)
}

func (q *Queue) popBacklog() (Job, bool) {
	q.jobsMu.Lock()
	defer q.jobsMu.Unlock()

	if len(q.backlog) == 0 {
		return Job{}, false
	}

	ret := q.backlog[0]
	q.backlog = q.backlog[1:]

	q.ongoing[ret.album.ID] = ret

	return ret, true
}

func (q *Queue) lenBacklog() int {
	q.jobsMu.Lock()
	defer q.jobsMu.Unlock()

	return len(q.backlog)
}

func (q *Queue) findAllAlbumsJobs() ([]Job, error) {
	log.Info(q.ctx, "find all job")
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
	log.Info(q.ctx, "find job for user", "user", user.ID)
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
