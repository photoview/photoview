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
		} else if mockScannerQueue.up_next[len(mockScannerQueue.up_next)-1].id != newJob.id {
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
	// A cancelled job winding down beside a restart of the same album: only
	// the live one belongs in the status.
	cancelledJob := makeScannerJob(100)
	cancelledJob.cancel()

	mockScannerQueue := ScannerQueue{
		idle_chan:   make(chan bool, 1),
		in_progress: []ScannerJob{cancelledJob, makeScannerJob(100)},
		up_next:     []ScannerJob{makeScannerJob(20), makeScannerJob(42)},
		db:          nil,
	}

	items := mockScannerQueue.GetQueueStatus()
	if len(items) != 3 {
		t.Fatalf("Expected 3 queue items, got %d", len(items))
	}

	if items[0].Album.ID != 100 || items[0].Status != models.ScannerJobStatusRunning {
		t.Errorf("Expected the in-progress album first and marked running, got %+v", items[0])
	}
	for _, item := range items[1:] {
		if item.Status != models.ScannerJobStatusQueued {
			t.Errorf("Expected album %d to be queued, got %s", item.Album.ID, item.Status)
		}
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

		if !mockScannerQueue.CancelJob(20) {
			t.Fatal("Expected CancelJob to return true for a queued job")
		}

		if len(mockScannerQueue.up_next) != 1 || mockScannerQueue.up_next[0].ctx.GetAlbum().ID != 42 {
			t.Errorf("Expected the queued job to be removed, up_next: %+v", mockScannerQueue.up_next)
		}

		if queuedJob.ctx.Err() == nil {
			t.Error("Expected the cancelled job's context to report an error")
		}
	})

	t.Run("cancelling a running job leaves it in place until it exits", func(t *testing.T) {
		runningJob := makeScannerJob(100)
		mockScannerQueue := ScannerQueue{
			idle_chan:   make(chan bool, 1),
			in_progress: []ScannerJob{runningJob},
			up_next:     make([]ScannerJob, 0),
			db:          nil,
		}

		if !mockScannerQueue.CancelJob(100) {
			t.Fatal("Expected CancelJob to return true for a running job")
		}

		if len(mockScannerQueue.in_progress) != 1 {
			t.Errorf("Expected the running job to stay until it exits on its own, got %+v", mockScannerQueue.in_progress)
		}
		if runningJob.ctx.Err() == nil {
			t.Error("Expected the cancelled job's context to report an error")
		}
	})

	t.Run("cancelling skips an already-cancelled job for the same album", func(t *testing.T) {
		// jobOnQueue lets a restart start while the cancelled job is still
		// finishing its current file, so both are in_progress for a while.
		cancelledJob := makeScannerJob(100)
		cancelledJob.cancel()
		restartedJob := makeScannerJob(100)

		mockScannerQueue := ScannerQueue{
			idle_chan:   make(chan bool, 1),
			in_progress: []ScannerJob{cancelledJob, restartedJob},
			up_next:     make([]ScannerJob, 0),
			db:          nil,
		}

		if !mockScannerQueue.CancelJob(100) {
			t.Fatal("Expected CancelJob to return true for the still-running restart")
		}

		if restartedJob.ctx.Err() == nil {
			t.Error("Expected the restarted job to be the one that got cancelled")
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

func TestScannerQueueCancelAllJobs(t *testing.T) {
	t.Run("cancels queued and running jobs and reports how many", func(t *testing.T) {
		mockScannerQueue := ScannerQueue{
			idle_chan:   make(chan bool, 1),
			in_progress: []ScannerJob{makeScannerJob(100)},
			up_next:     []ScannerJob{makeScannerJob(20), makeScannerJob(42)},
			db:          nil,
		}

		if cancelled := mockScannerQueue.CancelAllJobs(); cancelled != 3 {
			t.Errorf("Expected 3 jobs to be cancelled, got %d", cancelled)
		}

		if len(mockScannerQueue.up_next) != 0 {
			t.Errorf("Expected up_next to be emptied, got %+v", mockScannerQueue.up_next)
		}
		if len(mockScannerQueue.in_progress) != 1 {
			t.Errorf("Expected in_progress to stay until jobs exit on their own, got %+v", mockScannerQueue.in_progress)
		}
	})

	t.Run("does not count a job that was already cancelled", func(t *testing.T) {
		cancelledJob := makeScannerJob(100)
		cancelledJob.cancel()

		mockScannerQueue := ScannerQueue{
			idle_chan:   make(chan bool, 1),
			in_progress: []ScannerJob{cancelledJob, makeScannerJob(42)},
			up_next:     make([]ScannerJob, 0),
			db:          nil,
		}

		if cancelled := mockScannerQueue.CancelAllJobs(); cancelled != 1 {
			t.Errorf("Expected only the still-running job to be counted, got %d", cancelled)
		}
	})
}

func TestScannerQueueOneLiveJobPerAlbum(t *testing.T) {
	t.Run("a running job keeps a second job for the same album off the queue", func(t *testing.T) {
		mockScannerQueue := ScannerQueue{
			idle_chan:   make(chan bool, 1),
			in_progress: []ScannerJob{makeScannerJob(100)},
			up_next:     make([]ScannerJob, 0),
			db:          nil,
		}

		duplicate := makeScannerJob(100)
		if err := mockScannerQueue.addJob(&duplicate); err != nil {
			t.Fatalf("addJob returned an error: %v", err)
		}

		// Without this, a queued and a running job would share an album and
		// there would be no telling which one a cancel refers to.
		if len(mockScannerQueue.up_next) != 0 {
			t.Errorf("Expected the duplicate to be rejected, up_next: %+v", mockScannerQueue.up_next)
		}
	})

	t.Run("a cancelled running job does not block a restart", func(t *testing.T) {
		cancelledJob := makeScannerJob(100)
		cancelledJob.cancel()

		mockScannerQueue := ScannerQueue{
			idle_chan:   make(chan bool, 1),
			in_progress: []ScannerJob{cancelledJob},
			up_next:     make([]ScannerJob, 0),
			db:          nil,
		}

		restart := makeScannerJob(100)
		if err := mockScannerQueue.addJob(&restart); err != nil {
			t.Fatalf("addJob returned an error: %v", err)
		}

		if len(mockScannerQueue.up_next) != 1 {
			t.Fatalf("Expected the restart to be queued, up_next: %+v", mockScannerQueue.up_next)
		}

		// The album is listed once, not twice: the dead job is on its way out
		// and has nothing left to report.
		status := mockScannerQueue.GetQueueStatus()
		if len(status) != 1 {
			t.Fatalf("Expected one item for the album, got %+v", status)
		}
		if status[0].Status != models.ScannerJobStatusQueued {
			t.Errorf("Expected the live restart to be the one listed, got %v", status[0].Status)
		}
	})
}
