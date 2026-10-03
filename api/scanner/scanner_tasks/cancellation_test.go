package scanner_tasks

import (
	"context"
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/media_encoding"
	"github.com/photoview/photoview/api/scanner/scanner_task"
	// Registers -database and -filesystem, which CI passes to every package's
	// tests; a test binary that does not know them refuses to start.
	_ "github.com/photoview/photoview/api/test_utils/flags"
)

// cancelDuringProcessing cancels the scan job from inside a file's pipeline,
// as a cancelScanJob request arriving mid-file would.
type cancelDuringProcessing struct {
	scanner_task.ScannerTaskBase
	cancel context.CancelFunc
}

func (t cancelDuringProcessing) ProcessMedia(scanner_task.TaskContext, *media_encoding.EncodeMediaData, string) ([]*models.MediaURL, error) {
	t.cancel()

	return nil, nil
}

// recordProcessing notes that the pipeline reached it.
type recordProcessing struct {
	scanner_task.ScannerTaskBase
	ran *bool
}

func (t recordProcessing) ProcessMedia(scanner_task.TaskContext, *media_encoding.EncodeMediaData, string) ([]*models.MediaURL, error) {
	*t.ran = true

	return nil, nil
}

func runPipelineCancelledMidFile(t *testing.T, detach bool) (laterStepRan bool, err error) {
	t.Helper()

	previous := allTasks
	t.Cleanup(func() { allTasks = previous })

	jobCtx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	allTasks = []scanner_task.ScannerTask{
		cancelDuringProcessing{cancel: cancel},
		recordProcessing{ran: &laterStepRan},
	}

	ctx := scanner_task.NewTaskContext(jobCtx, nil, &models.Album{}, nil)
	if detach {
		ctx = ctx.WithoutCancel()
	}

	mediaData := media_encoding.NewEncodeMediaData(&models.Media{})
	_, err = Tasks.ProcessMedia(ctx, &mediaData, "")

	return laterStepRan, err
}

func TestCancellingAScanLetsTheCurrentFileFinish(t *testing.T) {
	// The documented contract: a running job stops once it finishes its
	// current file. ScanAlbum runs each file under a detached context for
	// exactly this - the pipeline checks for cancellation between its steps,
	// and a cancel arriving mid-file must not stop it there.
	laterStepRan, err := runPipelineCancelledMidFile(t, true)
	if err != nil {
		t.Fatalf("the file's pipeline failed after a mid-file cancel: %v", err)
	}

	if !laterStepRan {
		t.Error("the steps after the cancel never ran, so the file was left half processed")
	}
}

func TestTheJobContextAloneWouldStopMidFile(t *testing.T) {
	// Why the detached context is needed at all: under the job's own context
	// the same cancel stops the pipeline between two steps of one file.
	laterStepRan, err := runPipelineCancelledMidFile(t, false)
	if err == nil {
		t.Fatal("expected the job context's cancellation to stop the pipeline")
	}

	if laterStepRan {
		t.Error("expected the step after the cancel not to run under the job context")
	}
}
