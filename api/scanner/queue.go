package scanner

import (
	"database/sql"
	"errors"
	"sync"

	"github.com/viktorstrate/photoview/api/graphql/models"
)

type ScannerJobScope int

const (
	JOB_SCAN_USER ScannerJobScope = iota
	JOB_SCAN_ALBUM
)

type ScannerJob struct {
	scope ScannerJobScope
	// Either models.User, models.Album or nil depending on the value of scope
	model interface{}
}

func (job *ScannerJob) modelAsUser() (*models.User, error) {
	user, ok := job.model.(models.User)
	if !ok {
		return nil, errors.New("scanner job not of type User")
	}
	return &user, nil
}

func (job *ScannerJob) modelAsAlbum() (*models.Album, error) {
	album, ok := job.model.(models.Album)
	if !ok {
		return nil, errors.New("scanner job not of type Album")
	}
	return &album, nil
}

func (job *ScannerJob) Run() {
	// TODO: Not implemented
}

type ScannerQueue struct {
	mutex       sync.Mutex
	idle_chan   chan bool
	in_progress []ScannerJob
	up_next     []ScannerJob
	db          *sql.DB
}

var global_scanner_queue ScannerQueue

func InitializeScannerQueue(db *sql.DB) {
	global_scanner_queue = ScannerQueue{
		idle_chan:   make(chan bool, 1),
		in_progress: make([]ScannerJob, 0),
		up_next:     make([]ScannerJob, 0),
		db:          db,
	}
}

func (queue *ScannerQueue) startBackgroundWorker() {
	for {
		<-queue.idle_chan
		queue.mutex.Lock()
		defer queue.mutex.Unlock()
	}
}

func (queue *ScannerQueue) AddJob(job *ScannerJob) error {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	if exists, err := queue.jobOnQueue(job); exists || err != nil {
		return err
	}
	queue.up_next = append(queue.up_next, *job)
	queue.Notify()

	return nil
}

func (queue *ScannerQueue) Notify() bool {
	select {
	case queue.idle_chan <- true:
		return true
	default:
		return false
	}
}

func (queue *ScannerQueue) jobOnQueue(job *ScannerJob) (bool, error) {

	scannerJobs := append(queue.in_progress, queue.up_next...)

	for _, scannerJob := range scannerJobs {

		if scannerJob == *job {
			return true, nil
		}

		if scannerJob.scope == JOB_SCAN_USER {
			user, err := scannerJob.modelAsUser()
			if err != nil {
				return true, err
			}

			if job.scope == JOB_SCAN_ALBUM {
				album, err := job.modelAsAlbum()
				if err != nil {
					return true, err
				}

				if album.OwnerID == user.UserID {
					return true, nil
				}

			}
		}

	}

	return false, nil
}
