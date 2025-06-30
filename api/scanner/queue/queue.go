package queue

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/log"
	"github.com/photoview/photoview/api/scanner"
	"github.com/photoview/photoview/api/scanner/scanner_cache"
	"github.com/photoview/photoview/api/scanner/scanner_task"
	"github.com/photoview/photoview/api/scanner/scanner_utils"
	"gorm.io/gorm"
)

type queueJob struct {
	album *models.Album
	cache *scanner_cache.AlbumScannerCache
}

func (j *queueJob) Key() int {
	return j.album.ID
}

func (j *queueJob) String() string {
	return j.album.Title
}

type Queue struct {
	*commonQueue[*queueJob]

	ctx context.Context
	db  *gorm.DB
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

	ctx := log.WithAttrs(context.Background(), "process", "queue")

	ret := &Queue{
		ctx: ctx,
		db:  db,
	}
	commonQueue, err := newCommonQueue(ctx, interval, siteInfo.ConcurrentWorkers, ret)
	if err != nil {
		return nil, fmt.Errorf("can't create queue: %w", err)
	}
	ret.commonQueue = commonQueue

	return ret, nil
}

func (q *Queue) AddAllAlbums(ctx context.Context) error {
	jobs, err := q.findAllAlbumsJobs()
	if err != nil {
		return err
	}

	q.commonQueue.appendBacklog(jobs)

	return nil
}

func (q *Queue) AddUserAlbums(ctx context.Context, user *models.User) error {
	jobs, err := q.findUserAlbumsJobs(user)
	if err != nil {
		return fmt.Errorf("find albums for user (id: %d) error: %w", user.ID, err)
	}

	q.commonQueue.appendBacklog(jobs)

	return nil
}

// commonQueue callbacks
func (q *Queue) processJob(ctx context.Context, job *queueJob) {
	log.Info(ctx, "process album", "album", job.album.Title)

	task := scanner_task.NewTaskContext(ctx, q.db, job.album, job.cache)
	if err := scanner.ScanAlbum(task); err != nil {
		scanner_utils.ScannerError(ctx, "Failed to scan album: %v", err)
	}
}

func (q *Queue) periodicTrigger(ctx context.Context) {
	if err := q.AddAllAlbums(ctx); err != nil {
		log.Error(ctx, "fill queue error", "error", err)
	}
}

// helpers
func (q *Queue) findAllAlbumsJobs() ([]*queueJob, error) {
	log.Info(q.ctx, "find all job")
	var users []*models.User
	if err := q.db.Find(&users).Error; err != nil {
		return nil, fmt.Errorf("get all users from database error: %w", err)
	}

	var jobs []*queueJob

	for _, user := range users {
		job, err := q.findUserAlbumsJobs(user)
		if err != nil {
			return nil, fmt.Errorf("failed to add user (id: %d) for scanning: %w", user.ID, err)
		}
		jobs = append(jobs, job...)
	}

	return jobs, nil
}

func (q *Queue) findUserAlbumsJobs(user *models.User) ([]*queueJob, error) {
	log.Info(q.ctx, "find job for user", "user", user.ID)
	albumCache := scanner_cache.MakeAlbumCache()
	albums, album_errors := scanner.FindAlbumsForUser(q.db, user, albumCache)
	if err := errors.Join(album_errors...); err != nil {
		return nil, fmt.Errorf("find user(%d) album error: %w", user.ID, err)
	}

	jobs := make([]*queueJob, 0, len(albums))
	for _, album := range albums {
		jobs = append(jobs, &queueJob{
			album: album,
			cache: albumCache,
		})
	}

	return jobs, nil
}
