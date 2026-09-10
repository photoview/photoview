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

// TestScannerQueueJobOnQueueIgnoresCancelledInProgress covers a real bug: a
// cancelled in_progress job stays in that slice until it finishes its
// current file and exits on its own, so without this, restarting a scan
// right after cancelling it would find the album "still on queue" and be
// silently dropped - AddAlbumToQueue would report success while nothing was
// actually queued, for as long as the cancelled job's current file takes to
// finish.
func TestScannerQueueJobOnQueueIgnoresCancelledInProgress(t *testing.T) {
	cancelledJob := makeScannerJob(100)
	cancelledJob.cancel()

	mockScannerQueue := ScannerQueue{
		idle_chan:   make(chan bool, 1),
		in_progress: []ScannerJob{cancelledJob},
		up_next:     make([]ScannerJob, 0),
		db:          nil,
	}

	restart := makeScannerJob(100)
	onQueue, err := mockScannerQueue.jobOnQueue(&restart)
	if err != nil {
		t.Error("Expected jobOnQueue not to return an error")
	}
	if onQueue {
		t.Error("Expected a cancelled in_progress job not to block a restart of the same album")
	}

	if err := mockScannerQueue.addJob(&restart); err != nil {
		t.Errorf("addJob returned an unexpected error: %s", err)
	}
	if len(mockScannerQueue.up_next) != 1 {
		t.Errorf("Expected the restart to be queued in up_next, got %+v", mockScannerQueue.up_next)
	}
}

// TestScannerQueueRemoveFinishedJobMatchesByIdentity covers a real bug:
// jobOnQueue now lets a restart of an album be queued and promoted into
// in_progress while a cancelled job for the same album is still there
// finishing up, so in_progress can hold two entries with the same album ID.
// Removing "the job that just finished" by matching album ID could then
// remove the wrong entry - if the cancelled job finishes first but the
// removal matched the restart instead, the restart would be silently
// dropped from in_progress mid-run while the actually-finished cancelled
// job stayed behind forever, permanently occupying a worker slot.
func TestScannerQueueRemoveFinishedJobMatchesByIdentity(t *testing.T) {
	cancelledJob := makeScannerJob(100)
	restart := makeScannerJob(100)

	// The restart is listed first, so a removal that matched by album ID
	// (rather than by identity) would hit it before reaching the actually-
	// finished cancelledJob - the ordering here is what makes the test
	// actually discriminate between the two.
	mockScannerQueue := ScannerQueue{
		idle_chan:   make(chan bool, 1),
		in_progress: []ScannerJob{restart, cancelledJob},
		up_next:     make([]ScannerJob, 0),
		db:          nil,
	}

	// The cancelled job finishes first (it exits on its own once its
	// current file is done), while the restart for the same album is still
	// genuinely running.
	mockScannerQueue.removeFinishedJob(cancelledJob)

	if len(mockScannerQueue.in_progress) != 1 {
		t.Fatalf("Expected exactly one job left in_progress, got %d", len(mockScannerQueue.in_progress))
	}
	if mockScannerQueue.in_progress[0].id != restart.id {
		t.Errorf("Expected the restart to still be in_progress, but the wrong job was removed")
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

		ok := mockScannerQueue.CancelJob(100)
		if !ok {
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
		runningJob := makeScannerJob(100)
		queuedJobA := makeScannerJob(20)
		queuedJobB := makeScannerJob(42)

		mockScannerQueue := ScannerQueue{
			idle_chan:   make(chan bool, 1),
			in_progress: []ScannerJob{runningJob},
			up_next:     []ScannerJob{queuedJobA, queuedJobB},
			db:          nil,
		}

		cancelled := mockScannerQueue.CancelAllJobs()
		if cancelled != 3 {
			t.Errorf("Expected 3 jobs to be cancelled, got %d", cancelled)
		}

		if len(mockScannerQueue.up_next) != 0 {
			t.Errorf("Expected up_next to be emptied, got %+v", mockScannerQueue.up_next)
		}

		if len(mockScannerQueue.in_progress) != 1 {
			t.Errorf("Expected in_progress to stay in place until jobs exit on their own, got %+v", mockScannerQueue.in_progress)
		}

		for _, job := range []ScannerJob{runningJob, queuedJobA, queuedJobB} {
			if job.ctx.Err() == nil {
				t.Errorf("Expected job for album %d to have its context cancelled", job.ctx.GetAlbum().ID)
			}
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

	t.Run("an empty queue cancels nothing", func(t *testing.T) {
		mockScannerQueue := ScannerQueue{
			idle_chan:   make(chan bool, 1),
			in_progress: []ScannerJob{},
			up_next:     []ScannerJob{},
			db:          nil,
		}

		if cancelled := mockScannerQueue.CancelAllJobs(); cancelled != 0 {
			t.Errorf("Expected 0 jobs to be cancelled, got %d", cancelled)
		}
	})
}
