package scanner_queue

import (
	"context"
	"fmt"
	"github.com/photoview/photoview/api/scanner/media_type"
	"log"
	"path"
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

// ScannerJob describes a job on the queue to be run by the scanner over a single album
type ScannerJob struct {
	ctx scanner_task.TaskContext
	// album *models.Album
	// cache *scanner_cache.AlbumScannerCache
}

func NewScannerJob(ctx scanner_task.TaskContext) ScannerJob {
	return ScannerJob{
		ctx,
	}
}

func (job *ScannerJob) Run(db *gorm.DB) {
	err := scanner.ScanAlbum(job.ctx)
	if err != nil {
		scanner_utils.ScannerError("Failed to scan album: %v", err)
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
				if x.ctx == nextJob.ctx {
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

		if err := scanner.GenerateBlurhashes(queue.db); err != nil {
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

// AddUserToQueue finds all root albums owned by the given user and adds them to the scanner queue.
// Function does not block.
func AddUserToQueue(user *models.User) error {
	album_cache := scanner_cache.MakeAlbumCache()
	albums, album_errors := scanner.FindAlbumsForUser(global_scanner_queue.db, user, album_cache)
	for _, err := range album_errors {
		return errors.Wrapf(err, "find albums for user (user_id: %d)", user.ID)
	}

	global_scanner_queue.mutex.Lock()
	for _, album := range albums {
		global_scanner_queue.addJob(&ScannerJob{
			ctx: scanner_task.NewTaskContext(context.Background(), global_scanner_queue.db, album, make([]string, 0), album_cache),
		})
	}
	global_scanner_queue.mutex.Unlock()

	return nil
}

func AddMediaToQueue(mediaPath string) error {
	var media *models.Media

	mediatype, err := media_type.GetMediaType(mediaPath)
	if err != nil {
		return errors.Wrap(err, "not media")
	}
	if mediatype == nil || !mediatype.IsSupported() {
		return errors.New("unsupported media")
	}

	if err := global_scanner_queue.db.Preload("Album").Where("path = ?", mediaPath).Find(&media).Error; err != nil {
		return errors.Wrap(err, "media by path database query")
	}

	// add album to the queue
	var album *models.Album
	var parentAlbum *models.Album
	if media == nil || media.ID == 0 {

		albumPath := path.Dir(mediaPath)
		for album == nil {
			if err := global_scanner_queue.db.Where("path = ?", albumPath).Find(&album).Error; err != nil {
				return errors.Wrap(err, "album by path database query")
			}
			albumPath = path.Dir(albumPath)
		}
		if album == nil {
			return errors.New("No root album found")
		}
		if err := global_scanner_queue.db.Where("path = ?", albumPath).Find(&parentAlbum).Error; err != nil {
			return errors.Wrap(err, "parentalbum by path database query")
		}
	} else {
		album = &media.Album
	}

	var userAlbumOwner []*models.User
	if err := global_scanner_queue.db.Model(&album).Association("Owners").Find(&userAlbumOwner); err != nil {
		return errors.Wrap(err, "find owners for album")
	}

	if parentAlbum != nil && parentAlbum.ID == 0 {
		parentAlbum = nil
	}
	/*scanQueue := list.New()
	scanQueue.PushBack(scanner.ScanInfo{
		Path:   subalbumPath,
		Parent: parentAlbum,
		Ignore: nil,
	})
	//all_albums, album_errors := scanner.ProcessUserAlbums(scanQueue, global_scanner_queue.db, userAlbumOwner, album_cache)
	//for _, err := range album_errors {
	//	return errors.Wrapf(err, "find albums")
	//}*/

	album_cache := scanner_cache.MakeAlbumCache()
	album_cache.InsertAlbumIgnore(album.Path, []string{})
	var mediapaths []string
	if album.Path == path.Dir(mediaPath) {
		mediapaths = []string{path.Base(mediaPath)}
	} else {
		mediapaths = []string{}
	}
	global_scanner_queue.mutex.Lock()
	global_scanner_queue.addJob(&ScannerJob{
		ctx: scanner_task.NewTaskContext(context.Background(), global_scanner_queue.db, album, mediapaths, album_cache),
	})
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
