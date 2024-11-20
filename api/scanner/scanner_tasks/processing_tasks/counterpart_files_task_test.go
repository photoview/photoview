package processing_tasks

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/photoview/photoview/api/scanner/scanner_task"
	"github.com/photoview/photoview/api/test_utils/test_env"
	"github.com/photoview/photoview/api/utils"
)

func TestCounterpartFilesTaskMediaFound(t *testing.T) {
	tests := []struct {
		name                 string
		file                 string
		disableRawProcessing bool
		wantSkip             bool
	}{
		{
			name:                 "StandaloneProcessRaw",
			file:                 "standalone.jpg",
			disableRawProcessing: false,
			wantSkip:             false,
		},
		{
			name:                 "StandaloneNotProcessRaw",
			file:                 "standalone.jpg",
			disableRawProcessing: true,
			wantSkip:             false,
		},
		{
			name:                 "RawJpegProcessRaw",
			file:                 "fujifilm_raw.jpg",
			disableRawProcessing: false,
			wantSkip:             true,
		},
		{
			name:                 "RawJpegNotProcessRaw",
			file:                 "fujifilm_raw.jpg",
			disableRawProcessing: true,
			wantSkip:             false,
		},
		{
			name:                 "RawProcessRaw",
			file:                 "fujifilm_raw.raf",
			disableRawProcessing: false,
			wantSkip:             false,
		},
		{
			name:                 "RawNotProcessRaw",
			file:                 "fujifilm_raw.raf",
			disableRawProcessing: true,
			wantSkip:             true,
		},
		{
			name:                 "UnknownProcessRaw",
			file:                 "file.unknown",
			disableRawProcessing: false,
			wantSkip:             true,
		},
		{
			name:                 "UnknownNotProcessRaw",
			file:                 "file.unknown",
			disableRawProcessing: true,
			wantSkip:             true,
		},
	}

	mediaPath := test_env.PathFromAPIRoot("scanner/test_data/fake_media")

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			done := test_env.SetEnv(string(utils.EnvDisableRawProcessing), fmt.Sprintf("%v", tc.disableRawProcessing))
			defer done()

			ctx := scanner_task.NewTaskContext(context.Background(), nil, nil, nil)

			fname := filepath.Join(mediaPath, tc.file)
			fi, err := os.Stat(fname)
			if err != nil {
				t.Fatalf("Stat(%q) error: %v", fname, err)
			}

			var task CounterpartFilesTask
			got, err := task.MediaFound(ctx, fi, fname)
			if err != nil {
				t.Fatalf("task.MediaFound(ctx, %q) error: %v", fname, err)
			}

			if got, want := got, tc.wantSkip; got != want {
				t.Errorf("task.MediaFound(ctx, %q) = (skip)%v, want skip: %v", fname, got, want)
			}
		})
	}
}
