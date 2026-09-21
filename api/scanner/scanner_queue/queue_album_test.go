package scanner_queue

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// useTestQueue points the process-wide queue at the test database without
// starting its background worker, so the jobs stay where the test can read
// them instead of being scanned.
func useTestQueue(t *testing.T, db *gorm.DB) {
	t.Helper()

	global_scanner_queue = ScannerQueue{
		idle_chan:   make(chan bool, 1),
		in_progress: make([]ScannerJob, 0),
		up_next:     make([]ScannerJob, 0),
		db:          db,
	}
	t.Cleanup(func() { global_scanner_queue = ScannerQueue{} })
}

func queuedPaths() []string {
	paths := make([]string, 0, len(global_scanner_queue.up_next))
	for _, job := range global_scanner_queue.up_next {
		paths = append(paths, job.ctx.GetAlbum().Path)
	}
	sort.Strings(paths)

	return paths
}

func writePhoto(t *testing.T, dir string) {
	t.Helper()

	content, err := os.ReadFile("../test_media/orient/up_arrow_90cw_web.jpg")
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "photo.jpg"), content, 0o644))
}

func TestAddAlbumToQueueQueuesTheAlbumAndItsSubAlbumsOnly(t *testing.T) {
	db := test_utils.DatabaseTest(t)
	test_utils.FilesystemTest(t)
	useTestQueue(t, db)

	rootPath := t.TempDir()
	writePhoto(t, rootPath)
	writePhoto(t, filepath.Join(rootPath, "sub"))

	root := models.Album{Title: "root", Path: rootPath}
	require.NoError(t, db.Save(&root).Error)

	// Another album with photos on disk, outside the rescanned subtree.
	outsidePath := t.TempDir()
	writePhoto(t, outsidePath)
	require.NoError(t, db.Save(&models.Album{Title: "outside", Path: outsidePath}).Error)

	require.NoError(t, AddAlbumToQueue(&root))

	assert.Equal(t, []string{rootPath, filepath.Join(rootPath, "sub")}, queuedPaths())
}

func TestAddAlbumToQueueReportsAMissingDirectory(t *testing.T) {
	db := test_utils.DatabaseTest(t)
	test_utils.FilesystemTest(t)
	useTestQueue(t, db)

	gone := models.Album{Title: "gone", Path: filepath.Join(t.TempDir(), "gone")}
	require.NoError(t, db.Save(&gone).Error)

	assert.Error(t, AddAlbumToQueue(&gone))
	assert.Empty(t, queuedPaths(), "nothing is queued for an album that cannot be walked")
}
