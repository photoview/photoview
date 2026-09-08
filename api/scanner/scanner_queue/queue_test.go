package scanner_queue

import (
	"context"
	"flag"
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/scanner_cache"
	"github.com/photoview/photoview/api/scanner/scanner_task"
)

var _ = flag.Bool("database", false, "run database integration tests")
var _ = flag.Bool("filesystem", false, "run filesystem integration tests")

func makeAlbumWithID(id int) *models.Album {
	var album models.Album
	album.ID = id

	return &album
}

func makeScannerJob(albumID int) ScannerJob {
	ctx, cancel := context.WithCancel(context.Background())
	return NewScannerJob(scanner_task.NewTaskContext(ctx, nil, makeAlbumWithID(albumID), scanner_cache.MakeAlbumCache()), cancel)
}

func TestScannerQueueAddJob(t *testing.T) {

	scannerJobs := []ScannerJob{
		makeScannerJob(100),
		makeScannerJob(20),
	}

	mockScannerQueue := ScannerQueue{
		idle_chan:   make(chan bool, 1),
		in_progress: make([]ScannerJob, 0),
		up_next:     scannerJobs,
		db:          nil,
	}

	t.Run("add new job to scanner queue", func(t *testing.T) {
		newJob := makeScannerJob(42)

		startingJobs := len(mockScannerQueue.up_next)

		err := mockScannerQueue.addJob(&newJob)
		if err != nil {
			t.Errorf(".AddJob() returned an unexpected error: %s", err)
		}

		if len(mockScannerQueue.up_next) != startingJobs+1 {
			t.Errorf("Expected scanner queue length to be %d but got %d", startingJobs+1, len(mockScannerQueue.up_next))
		} else if mockScannerQueue.up_next[len(mockScannerQueue.up_next)-1].ctx.GetAlbum().ID != newJob.ctx.GetAlbum().ID {
			t.Errorf("Expected scanner queue to contain the job that was added: %+v", newJob)
		}

	})

	t.Run("add existing job to scanner queue", func(t *testing.T) {
		startingJobs := len(mockScannerQueue.up_next)

		job := makeScannerJob(20)
		err := mockScannerQueue.addJob(&job)
		if err != nil {
			t.Errorf(".AddJob() returned an unexpected error: %s", err)
		}

		if len(mockScannerQueue.up_next) != startingJobs {
			t.Errorf("Expected scanner queue length not to change: start length %d, new length %d", startingJobs, len(mockScannerQueue.up_next))
		}

	})
}

func TestScannerQueueJobOnQueue(t *testing.T) {

	scannerJobs := []ScannerJob{
		makeScannerJob(100),
		makeScannerJob(20),
	}

	mockScannerQueue := ScannerQueue{
		idle_chan:   make(chan bool, 1),
		in_progress: make([]ScannerJob, 0),
		up_next:     scannerJobs,
		db:          nil,
	}

	onQueueTests := []struct {
		string
		bool
		ScannerJob
	}{
		{"album which owner is already on the queue", true, makeScannerJob(100)},
		{"album that is not on the queue", false, makeScannerJob(321)},
	}

	for _, test := range onQueueTests {
		t.Run(test.string, func(t *testing.T) {
			onQueue, err := mockScannerQueue.jobOnQueue(&test.ScannerJob)
			if err != nil {
				t.Error("Expected jobOnQueue not to return an error")
			} else if onQueue != test.bool {
				t.Fail()
			}
		})
	}

}

func TestScannerQueueGetQueueStatus(t *testing.T) {

	mockScannerQueue := ScannerQueue{
		idle_chan:   make(chan bool, 1),
		in_progress: []ScannerJob{makeScannerJob(100)},
		up_next:     []ScannerJob{makeScannerJob(20), makeScannerJob(42)},
		db:          nil,
	}

	items := mockScannerQueue.GetQueueStatus()

	if len(items) != 3 {
		t.Fatalf("Expected 3 queue items but got %d", len(items))
	}

	if items[0].Album.ID != 100 || items[0].Status != models.ScannerJobStatusRunning {
		t.Errorf("Expected first item to be running album 100, got album %d with status %s", items[0].Album.ID, items[0].Status)
	}

	if items[1].Album.ID != 20 || items[1].Status != models.ScannerJobStatusQueued {
		t.Errorf("Expected second item to be queued album 20, got album %d with status %s", items[1].Album.ID, items[1].Status)
	}

	if items[2].Album.ID != 42 || items[2].Status != models.ScannerJobStatusQueued {
		t.Errorf("Expected third item to be queued album 42, got album %d with status %s", items[2].Album.ID, items[2].Status)
	}
}

func TestScannerQueueCancelJob(t *testing.T) {

	t.Run("cancelling a queued job removes it and cancels its context", func(t *testing.T) {
		queuedJob := makeScannerJob(20)
		mockScannerQueue := ScannerQueue{
			idle_chan:   make(chan bool, 1),
			in_progress: []ScannerJob{makeScannerJob(100)},
			up_next:     []ScannerJob{queuedJob, makeScannerJob(42)},
			db:          nil,
		}

		ok := mockScannerQueue.CancelJob(20)
		if !ok {
			t.Fatal("Expected CancelJob to return true for a queued job")
		}

		if len(mockScannerQueue.up_next) != 1 || mockScannerQueue.up_next[0].ctx.GetAlbum().ID != 42 {
			t.Errorf("Expected the queued job to be removed, up_next: %+v", mockScannerQueue.up_next)
		}

		if len(mockScannerQueue.in_progress) != 1 {
			t.Errorf("Expected in_progress to be untouched, got %+v", mockScannerQueue.in_progress)
		}

		if queuedJob.ctx.Err() == nil {
			t.Error("Expected the cancelled job's context to report an error")
		}
	})

	t.Run("cancelling a running job cancels its context but leaves it in place", func(t *testing.T) {
		runningJob := makeScannerJob(100)
		mockScannerQueue := ScannerQueue{
			idle_chan:   make(chan bool, 1),
			in_progress: []ScannerJob{runningJob},
			up_next:     []ScannerJob{makeScannerJob(20)},
			db:          nil,
		}

		ok := mockScannerQueue.CancelJob(100)
		if !ok {
			t.Fatal("Expected CancelJob to return true for a running job")
		}

		if len(mockScannerQueue.in_progress) != 1 {
			t.Errorf("Expected the running job to stay in in_progress until it exits on its own, got %+v", mockScannerQueue.in_progress)
		}

		if runningJob.ctx.Err() == nil {
			t.Error("Expected the cancelled job's context to report an error")
		}
	})

	t.Run("cancelling an unknown album returns false", func(t *testing.T) {
		mockScannerQueue := ScannerQueue{
			idle_chan:   make(chan bool, 1),
			in_progress: []ScannerJob{makeScannerJob(100)},
			up_next:     []ScannerJob{makeScannerJob(20)},
			db:          nil,
		}

		if mockScannerQueue.CancelJob(999) {
			t.Error("Expected CancelJob to return false for an album with no matching job")
		}
	})
}
