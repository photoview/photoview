package scanner

import (
	"database/sql"
	"log"
	"sync"

	"github.com/viktorstrate/photoview/api/graphql/models"
)

type ScannerJob struct {
	album *models.Album
	cache *AlbumScannerCache
}

func (job *ScannerJob) Run(db *sql.DB) {
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
	db          *sql.DB
	settings    ScannerQueueSettings
}

var global_scanner_queue ScannerQueue

func InitializeScannerQueue(db *sql.DB) {
	global_scanner_queue = ScannerQueue{
		idle_chan:   make(chan bool, 1),
		in_progress: make([]ScannerJob, 0),
		up_next:     make([]ScannerJob, 0),
		db:          db,
		settings:    ScannerQueueSettings{max_concurrent_tasks: 3},
	}

	go global_scanner_queue.startBackgroundWorker()
}

func (queue *ScannerQueue) startBackgroundWorker() {
	for {
		log.Println("Queue waiting")
		<-queue.idle_chan
		log.Println("Queue waiting for lock")
		queue.mutex.Lock()
		log.Println("Queue running")

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

		log.Printf("Waiting jobs: %d\n", len(queue.up_next))

		queue.mutex.Unlock()
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

func AddUserToQueue(user *models.User) {
	album_cache := MakeAlbumCache()
	albums, album_errors := findAlbumsForUser(global_scanner_queue.db, user, album_cache)
	for _, err := range album_errors {
		log.Printf("User scanner error: %s", err)
	}

	global_scanner_queue.mutex.Lock()
	for _, album := range albums {
		global_scanner_queue.addJob(&ScannerJob{
			album: album,
			cache: album_cache,
		})
	}
	global_scanner_queue.mutex.Unlock()
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
		if scannerJob.album.AlbumID == job.album.AlbumID {
			return true, nil
		}
	}

	return false, nil
}
