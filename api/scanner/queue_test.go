package scanner

import (
	"testing"

	"github.com/viktorstrate/photoview/api/graphql/models"
)

func TestScannerQueue_AddJob(t *testing.T) {

	scannerJobs := []ScannerJob{
		{album: &models.Album{AlbumID: 100}, cache: MakeAlbumCache()},
		{album: &models.Album{AlbumID: 20}, cache: MakeAlbumCache()},
	}

	mockScannerQueue := ScannerQueue{
		idle_chan:   make(chan bool, 1),
		in_progress: make([]ScannerJob, 0),
		up_next:     scannerJobs,
		db:          nil,
	}

	t.Run("add new job to scanner queue", func(t *testing.T) {
		newJob := ScannerJob{album: &models.Album{AlbumID: 42}, cache: MakeAlbumCache()}

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

		err := mockScannerQueue.addJob(&ScannerJob{album: &models.Album{AlbumID: 20}, cache: MakeAlbumCache()})
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
		{album: &models.Album{AlbumID: 100}, cache: MakeAlbumCache()},
		{album: &models.Album{AlbumID: 20}, cache: MakeAlbumCache()},
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
			album: &models.Album{AlbumID: 100}, cache: MakeAlbumCache(),
		}},
		{"album that is not on the queue", false, ScannerJob{
			album: &models.Album{AlbumID: 321}, cache: MakeAlbumCache(),
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
