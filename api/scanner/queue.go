package scanner

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/graphql/notification"
	"github.com/photoview/photoview/api/scanner/scanner_cache"
	"github.com/photoview/photoview/api/scanner/scanner_utils"
	"github.com/photoview/photoview/api/utils"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type ScannerJob struct {
	album *models.Album
	cache *scanner_cache.AlbumScannerCache
}

func (job *ScannerJob) Run(db *gorm.DB) {
	scanAlbum(job.album, job.cache, db)
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
		should_stop := queue.close_chan != nil && len(queue.in_progress) == 0 && len(queue.up_next) == 0
		queue.running = false
		queue.mutex.Unlock()

		if should_stop {
			*queue.close_chan <- true
			break
		}

		queue.processQueue(&notifyThrottle)
	}

	log.Println("Scanner background worker stopped")
}

func (queue *ScannerQueue) CloseBackgroundWorker() {
	queue.mutex.Lock()
	close_chan := make(chan bool)
	queue.close_chan = &close_chan
	queue.mutex.Unlock()

	queue.notify()

	log.Println("Waiting for scanner background worker to finish all jobs...")
	<-close_chan
}

func (queue *ScannerQueue) processQueue(notifyThrottle *utils.Throttle) {
	log.Println("Queue waiting for lock")
	queue.mutex.Lock()
	log.Printf("Queue running: in_progress: %d, max_tasks: %d, queue_len: %d\n", len(queue.in_progress), queue.settings.max_concurrent_tasks, len(queue.up_next))

	for len(queue.in_progress) < queue.settings.max_concurrent_tasks && len(queue.up_next) > 0 {
		log.Println("Queue starting job")
		nextJob := queue.up_next[0]
		queue.up_next = queue.up_next[1:]
		queue.in_progress = append(queue.in_progress, nextJob)

		go func() {
			log.Println("Starting job")
			nextJob.Run(queue.db)
			log.Println("Job finished")

			// Delete finished job from queue
			queue.mutex.Lock()
			for i, x := range queue.in_progress {
				if x == nextJob {
					queue.in_progress[i] = queue.in_progress[len(queue.in_progress)-1]
					queue.in_progress = queue.in_progress[0 : len(queue.in_progress)-1]
					break
				}
			}
			queue.mutex.Unlock()

			queue.notify()
		}()
	}

	in_progress_length := len(global_scanner_queue.in_progress)
	up_next_length := len(global_scanner_queue.up_next)

	queue.mutex.Unlock()

	if in_progress_length+up_next_length == 0 {
		notification.BroadcastNotification(&models.Notification{
			Key:      "global-scanner-progress",
			Type:     models.NotificationTypeMessage,
			Header:   "Generating blurhashes",
			Content:  "Generating blurhashes for newly scanned media",
			Positive: true,
		})

		if err := GenerateBlurhashes(queue.db); err != nil {
			scanner_utils.ScannerError("Failed to generate blurhashes: %v", err)
		}

		notification.BroadcastNotification(&models.Notification{
			Key:      "global-scanner-progress",
			Type:     models.NotificationTypeMessage,
			Header:   "Scanner complete",
			Content:  "All jobs have been scanned",
			Positive: true,
		})
	} else {
		notifyThrottle.Trigger(func() {
			notification.BroadcastNotification(&models.Notification{
				Key:     "global-scanner-progress",
				Type:    models.NotificationTypeMessage,
				Header:  "Scanning media",
				Content: fmt.Sprintf("%d jobs in progress\n%d jobs waiting", in_progress_length, up_next_length),
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

func AddUserToQueue(user *models.User) error {
	album_cache := scanner_cache.MakeAlbumCache()
	albums, album_errors := findAlbumsForUser(global_scanner_queue.db, user, album_cache)
	for _, err := range album_errors {
		return errors.Wrapf(err, "find albums for user (user_id: %d)", user.ID)
	}

	global_scanner_queue.mutex.Lock()
	for _, album := range albums {
		global_scanner_queue.addJob(&ScannerJob{
			album: album,
			cache: album_cache,
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
		if scannerJob.album.ID == job.album.ID {
			return true, nil
		}
	}

	return false, nil
}
