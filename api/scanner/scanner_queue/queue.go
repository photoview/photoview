package scanner_queue

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/graphql/notification"
	"github.com/photoview/photoview/api/scanner"
	"github.com/photoview/photoview/api/scanner/scanner_cache"
	"github.com/photoview/photoview/api/scanner/scanner_task"
	"github.com/photoview/photoview/api/scanner/scanner_utils"
	"github.com/photoview/photoview/api/utils"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const globalScannerProgress = "global-scanner-progress"

// ScannerJob describes a job on the queue to be run by the scanner over a single album
type ScannerJob struct {
	ctx scanner_task.TaskContext
	// cancel stops this job: an up_next job is simply dropped, an
	// in_progress job finishes processing its current file and then exits
	// early instead of continuing to the next one.
	cancel context.CancelFunc
	// album *models.Album
	// cache *scanner_cache.AlbumScannerCache
}

func NewScannerJob(ctx scanner_task.TaskContext, cancel context.CancelFunc) ScannerJob {
	return ScannerJob{
		ctx,
		cancel,
	}
}

func (job *ScannerJob) Run(db *gorm.DB) {
	err := scanner.ScanAlbum(job.ctx)
	if err != nil && !errors.Is(err, context.Canceled) {
		scanner_utils.ScannerError(nil, "Failed to scan album: %v", err)
	}
}

type ScannerQueueSettings struct {
	max_concurrent_tasks int
}

type ScannerQueue struct {
	mutex       sync.Mutex
	idle_chan   chan bool
	in_progress []ScannerJob
	up_next     []ScannerJob
	db          *gorm.DB
	settings    ScannerQueueSettings
	close_chan  *chan bool
	running     bool
}

var global_scanner_queue ScannerQueue

func InitializeScannerQueue(db *gorm.DB) error {

	var concurrentWorkers int
	{
		site_info, err := models.GetSiteInfo(db)
		if err != nil {
			return errors.Wrap(err, "get current workers from database")
		}
		concurrentWorkers = site_info.ConcurrentWorkers
	}

	log.Printf("Initializing scanner queue with %d workers", concurrentWorkers)

	global_scanner_queue = ScannerQueue{
		idle_chan:   make(chan bool, 1),
		in_progress: make([]ScannerJob, 0),
		up_next:     make([]ScannerJob, 0),
		db:          db,
		settings:    ScannerQueueSettings{max_concurrent_tasks: concurrentWorkers},
		close_chan:  nil,
		running:     true,
	}

	go global_scanner_queue.startBackgroundWorker()

	return nil
}

func CloseScannerQueue() {
	global_scanner_queue.CloseBackgroundWorker()
}

func ChangeScannerConcurrentWorkers(newMaxWorkers int) {
	global_scanner_queue.mutex.Lock()
	defer global_scanner_queue.mutex.Unlock()

	log.Printf("Scanner max concurrent workers changed to: %d", newMaxWorkers)
	global_scanner_queue.settings.max_concurrent_tasks = newMaxWorkers
}

func (queue *ScannerQueue) startBackgroundWorker() {

	notifyThrottle := utils.NewThrottle(500 * time.Millisecond)

	for {
		log.Println("Queue waiting")
		<-queue.idle_chan

		queue.mutex.Lock()
		shouldStop := queue.close_chan != nil && len(queue.in_progress) == 0 && len(queue.up_next) == 0
		queue.running = false
		queue.mutex.Unlock()

		if shouldStop {
			*queue.close_chan <- true
			break
		}

		queue.processQueue(&notifyThrottle)
	}

	log.Println("Scanner background worker stopped")
}

func (queue *ScannerQueue) CloseBackgroundWorker() {
	queue.mutex.Lock()
	closeChan := make(chan bool)
	queue.close_chan = &closeChan
	queue.mutex.Unlock()

	queue.notify()

	log.Println("Waiting for scanner background worker to finish all jobs...")
	<-closeChan
}

func (queue *ScannerQueue) processQueue(notifyThrottle *utils.Throttle) {
	log.Println("Queue waiting for lock")
	queue.mutex.Lock()
	maxJobs := queue.settings.max_concurrent_tasks
	log.Printf("Queue running: in_progress: %d, max_tasks: %d, queue_len: %d\n", len(queue.in_progress), maxJobs, len(queue.up_next))

	for len(queue.in_progress) < maxJobs && len(queue.up_next) > 0 {
		log.Println("Queue starting job")
		nextJob := queue.up_next[0]
		queue.up_next = queue.up_next[1:]
		queue.in_progress = append(queue.in_progress, nextJob)
		jobNum := len(queue.in_progress)

		go func() {
			log.Printf("Starting job %d/%d\n", jobNum, maxJobs)
			nextJob.Run(queue.db)
			log.Printf("Finished job %d/%d\n", jobNum, maxJobs)

			// Delete finished job from queue
			queue.mutex.Lock()
			for i, x := range queue.in_progress {
				if x.ctx.GetAlbum().ID == nextJob.ctx.GetAlbum().ID {
					queue.in_progress[i] = queue.in_progress[len(queue.in_progress)-1]
					queue.in_progress = queue.in_progress[0 : len(queue.in_progress)-1]
					break
				}
			}
			queue.mutex.Unlock()

			queue.notify()
		}()
	}

	inProgressLength := len(global_scanner_queue.in_progress)
	upNextLength := len(global_scanner_queue.up_next)

	queue.mutex.Unlock()

	if inProgressLength+upNextLength == 0 {
		notification.BroadcastNotification(&models.Notification{
			Key:      globalScannerProgress,
			Type:     models.NotificationTypeMessage,
			Header:   "Scanner complete",
			Content:  "All jobs have been scanned",
			Positive: true,
		})
	} else {
		notifyThrottle.Trigger(func() {
			notification.BroadcastNotification(&models.Notification{
				Key:     globalScannerProgress,
				Type:    models.NotificationTypeMessage,
				Header:  "Scanning media",
				Content: fmt.Sprintf("%d jobs in progress\n%d jobs waiting", inProgressLength, upNextLength),
			})
		})
	}
}

// Notifies the queue that the jobs has changed
func (queue *ScannerQueue) notify() bool {
	select {
	case queue.idle_chan <- true:
		return true
	default:
		return false
	}
}

func AddAllToQueue() error {

	var users []*models.User
	result := global_scanner_queue.db.Find(&users)
	if result.Error != nil {
		return errors.Wrap(result.Error, "get all users from database")
	}

	for _, user := range users {
		if err := AddUserToQueue(user); err != nil {
			return errors.Wrapf(err, "failed to add user for scanning (%d)", user.ID)
		}
	}

	return nil
}

// AddUserToQueue finds all root albums owned by the given user and adds them to the scanner queue.
// Function does not block.
func AddUserToQueue(user *models.User) error {
	albumCache := scanner_cache.MakeAlbumCache()
	albums, album_errors := scanner.FindAlbumsForUser(global_scanner_queue.db, user, albumCache)
	for _, err := range album_errors {
		return errors.Wrapf(err, "find albums for user (user_id: %d)", user.ID)
	}

	global_scanner_queue.mutex.Lock()
	for _, album := range albums {
		jobCtx, cancel := context.WithCancel(context.Background())
		global_scanner_queue.addJob(&ScannerJob{
			ctx:    scanner_task.NewTaskContext(jobCtx, global_scanner_queue.db, album, albumCache),
			cancel: cancel,
		})
	}
	global_scanner_queue.mutex.Unlock()

	return nil
}

// AddAlbumToQueue recursively finds the given album and its sub-albums and adds
// them to the scanner queue, without touching the rest of the library.
// Function does not block.
func AddAlbumToQueue(album *models.Album) error {
	albumCache := scanner_cache.MakeAlbumCache()
	albums, album_errors := scanner.FindAlbumsForAlbum(global_scanner_queue.db, album, albumCache)
	for _, err := range album_errors {
		return errors.Wrapf(err, "find albums for album (album_id: %d)", album.ID)
	}

	global_scanner_queue.mutex.Lock()
	for _, album := range albums {
		jobCtx, cancel := context.WithCancel(context.Background())
		global_scanner_queue.addJob(&ScannerJob{
			ctx:    scanner_task.NewTaskContext(jobCtx, global_scanner_queue.db, album, albumCache),
			cancel: cancel,
		})
	}
	global_scanner_queue.mutex.Unlock()

	return nil
}

// Queue should be locked prior to calling this function
func (queue *ScannerQueue) addJob(job *ScannerJob) error {
	if exists, err := queue.jobOnQueue(job); exists || err != nil {
		return err
	}
	queue.up_next = append(queue.up_next, *job)
	queue.notify()

	return nil
}

// Queue should be locked prior to calling this function
func (queue *ScannerQueue) jobOnQueue(job *ScannerJob) (bool, error) {

	scannerJobs := append(queue.in_progress, queue.up_next...)

	for _, scannerJob := range scannerJobs {
		if scannerJob.ctx.GetAlbum().ID == job.ctx.GetAlbum().ID {
			return true, nil
		}
	}

	return false, nil
}

// GetQueueStatus returns a snapshot of the jobs currently running or
// waiting on this queue, one entry per album/sub-album - most useful for
// surfacing "what's still left to scan" without waiting on the generic
// broadcast notifications.
func (queue *ScannerQueue) GetQueueStatus() []models.ScannerQueueItem {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	items := make([]models.ScannerQueueItem, 0, len(queue.in_progress)+len(queue.up_next))
	for _, job := range queue.in_progress {
		items = append(items, models.ScannerQueueItem{
			Album:  job.ctx.GetAlbum(),
			Status: models.ScannerJobStatusRunning,
		})
	}
	for _, job := range queue.up_next {
		items = append(items, models.ScannerQueueItem{
			Album:  job.ctx.GetAlbum(),
			Status: models.ScannerJobStatusQueued,
		})
	}
	return items
}

// GetQueueStatus returns a snapshot of the global scanner queue's jobs.
func GetQueueStatus() []models.ScannerQueueItem {
	return global_scanner_queue.GetQueueStatus()
}

// CancelJob cancels a specific album's scanner job, whether queued or
// currently running. A queued job is removed immediately; a running job is
// signalled to stop and finishes processing its current file before exiting,
// rather than being interrupted mid-file - it removes itself from
// in_progress on that natural exit, same as any other completed job.
// Returns false if no matching job was found.
func (queue *ScannerQueue) CancelJob(albumID int) bool {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	for i, job := range queue.up_next {
		if job.ctx.GetAlbum().ID == albumID {
			job.cancel()
			queue.up_next = append(queue.up_next[:i], queue.up_next[i+1:]...)
			return true
		}
	}

	for _, job := range queue.in_progress {
		if job.ctx.GetAlbum().ID == albumID {
			job.cancel()
			return true
		}
	}

	return false
}

// CancelJob cancels a specific album's job on the global scanner queue.
func CancelJob(albumID int) bool {
	return global_scanner_queue.CancelJob(albumID)
}

// CancelAllJobs cancels every job currently on the queue, whether queued or
// running, following the same semantics as CancelJob: queued jobs are
// removed immediately, running jobs are signalled to stop and finish
// processing their current file before exiting on their own. Returns the
// number of jobs cancelled.
func (queue *ScannerQueue) CancelAllJobs() int {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	cancelled := 0

	for _, job := range queue.up_next {
		job.cancel()
		cancelled++
	}
	queue.up_next = queue.up_next[:0]

	for _, job := range queue.in_progress {
		job.cancel()
		cancelled++
	}

	return cancelled
}

// CancelAllJobs cancels every job on the global scanner queue.
func CancelAllJobs() int {
	return global_scanner_queue.CancelAllJobs()
}
