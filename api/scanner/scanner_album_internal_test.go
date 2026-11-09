package scanner

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"testing"

	"github.com/photoview/photoview/api/scanner/media_type"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorthRetrying(t *testing.T) {
	// A file that vanished between the directory listing and the scan.
	_, missingErr := os.Stat(path.Join(t.TempDir(), "gone.jpg"))
	require.Error(t, missingErr)

	cases := []struct {
		name  string
		err   error
		retry bool
	}{
		{
			name:  "a database error could pass on a second attempt",
			err:   errors.New("Error 1205: Lock wait timeout exceeded"),
			retry: true,
		},
		{
			name:  "an unknown media type stays unknown",
			err:   fmt.Errorf("media %s: %w", "x.txt", media_type.ErrUnknownType),
			retry: false,
		},
		{
			name:  "an unknown media type from a task, wrapped again",
			err:   errors.Wrap(fmt.Errorf("sidecar: %w", media_type.ErrUnknownType), "scanning media error"),
			retry: false,
		},
		{
			name:  "a file that is gone stays gone",
			err:   errors.Wrap(missingErr, "scanning media error"),
			retry: false,
		},
		{
			name:  "a file that cannot be read stays unreadable",
			err:   errors.Wrap(fs.ErrPermission, "scanning media error"),
			retry: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.retry, worthRetrying(tc.err))
		})
	}
}
