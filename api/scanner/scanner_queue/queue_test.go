package scanner_queue

import (
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/scanner_cache"
)

func makeAlbumWithID(id int) *models.Album {
	var album models.Album
	album.ID = id

	return &album
}

func TestScannerQueue_AddJob(t *testing.T) {

	scannerJobs := []ScannerJob{
		{album: makeAlbumWithID(100), cache: scanner_cache.MakeAlbumCache()},
		{album: makeAlbumWithID(20), cache: scanner_cache.MakeAlbumCache()},
	}

	mockScannerQueue := ScannerQueue{
		idle_chan:   make(chan bool, 1),
		in_progress: make([]ScannerJob, 0),
		up_next:     scannerJobs,
		db:          nil,
	}

	t.Run("add new job to scanner queue", func(t *testing.T) {
		newJob := ScannerJob{album: makeAlbumWithID(42), cache: scanner_cache.MakeAlbumCache()}

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

		err := mockScannerQueue.addJob(&ScannerJob{album: makeAlbumWithID(20), cache: scanner_cache.MakeAlbumCache()})
		if err != nil {
			t.Errorf(".AddJob() returned an unexpected error: %s", err)
		}

		if len(mockScannerQueue.up_next) != startingJobs {
			t.Errorf("Expected scanner queue length not to change: start length %d, new length %d", startingJobs, len(mockScannerQueue.up_next))
		}

	})

}

func TestScannerQueue_JobOnQueue(t *testing.T) {

	scannerJobs := []ScannerJob{
		{album: makeAlbumWithID(100), cache: scanner_cache.MakeAlbumCache()},
		{album: makeAlbumWithID(20), cache: scanner_cache.MakeAlbumCache()},
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
		{"album which owner is already on the queue", true, ScannerJob{
			album: makeAlbumWithID(100), cache: scanner_cache.MakeAlbumCache(),
		}},
		{"album that is not on the queue", false, ScannerJob{
			album: makeAlbumWithID(321), cache: scanner_cache.MakeAlbumCache(),
		}},
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
