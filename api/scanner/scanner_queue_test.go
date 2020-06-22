package scanner

import (
	"testing"

	"github.com/viktorstrate/photoview/api/graphql/models"
)

func TestScannerQueue_JobOnQueue(t *testing.T) {

	scannerJobs := []ScannerJob{
		{scope: JOB_SCAN_ALBUM, model: models.Album{AlbumID: 100, OwnerID: 123}},
		{scope: JOB_SCAN_USER, model: models.User{UserID: 20}},
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
		{"user that is already on the queue", true, ScannerJob{
			scope: JOB_SCAN_USER,
			model: models.User{UserID: 20},
		}},
		{"album which owner is already on the queue", true, ScannerJob{
			scope: JOB_SCAN_ALBUM,
			model: models.Album{AlbumID: 40, OwnerID: 20},
		}},
		{"album that is not on the queue", false, ScannerJob{
			scope: JOB_SCAN_ALBUM,
			model: models.Album{AlbumID: 321, OwnerID: 11},
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
