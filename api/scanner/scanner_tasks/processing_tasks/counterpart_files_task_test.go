package processing_tasks_test

import (
	"context"
	"io/fs"
	"os"
	"testing"

	"github.com/photoview/photoview/api/scanner/scanner_task"
	"github.com/photoview/photoview/api/scanner/scanner_tasks/processing_tasks"
	"github.com/photoview/photoview/api/test_utils"
)

func TestCounterpartFilesTaskMediaFound(t *testing.T) {
	mediaPath := test_utils.PathFromAPIRoot("scanner/testdata/media")
	fileSys := os.DirFS(mediaPath)
	standalone, _ := fs.Stat(fileSys, "standalone.jpg")

	var task processing_tasks.CounterpartFilesTask
	skip, err := task.MediaFound(scanner_task.NewTaskContext(context.Background(), nil, nil, nil), standalone, mediaPath)
	if err != nil {
		t.Fatal("error:", err)
	}
	t.Error("skip:", skip)
}
