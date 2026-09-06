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
	return NewScannerJob(scanner_task.NewTaskContext(context.Background(), nil, makeAlbumWithID(albumID), scanner_cache.MakeAlbumCache()))
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
		} else if mockScannerQueue.up_next[len(mockScannerQueue.up_next)-1] != newJob {
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
