package queue

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/photoview/photoview/api/test_utils"
)

// testDataPath resolves a path under scanner/queue/test_data.
func testDataPath(parts ...string) string {
	return test_utils.PathFromAPIRoot(append([]string{"scanner", "queue", "test_data"}, parts...)...)
}

// copyFixtureFile copies a single file from scanner/queue/test_data/<relPath>
// to dst, so tests can freely process/mutate/delete their own copy without
// touching the checked-in fixture.
func copyFixtureFile(t *testing.T, relPath string, dst string) {
	t.Helper()

	src := testDataPath(relPath)
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("read fixture %s: %v", relPath, err)
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		t.Fatalf("write fixture copy to %s: %v", dst, err)
	}
}

// copyFixtureDir copies every file directly inside
// scanner/queue/test_data/<relDir> into dstDir, preserving filenames - used
// for fixture sets where the files must keep the same basename relationship
// to each other (e.g. raw_pair's photo.tiff/photo.jpg/photo.tiff.xmp).
func copyFixtureDir(t *testing.T, relDir string, dstDir string) {
	t.Helper()

	src := testDataPath(relDir)
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatalf("read fixture dir %s: %v", relDir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		copyFixtureFile(t, filepath.Join(relDir, entry.Name()), filepath.Join(dstDir, entry.Name()))
	}
}
