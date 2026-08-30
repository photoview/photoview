package queue

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/scanner_cache"
)

// TestExpandAlbumReadDirError checks that a missing album directory surfaces
// as an error, with no tasks/state built around it. expandAlbum never
// touches the database on this path, so no *gorm.DB is needed.
func TestExpandAlbumReadDirError(t *testing.T) {
	album := &models.Album{Title: "missing dir", Path: filepath.Join(t.TempDir(), "does-not-exist")}
	cache := scanner_cache.MakeAlbumCache()

	tasks, state, err := expandAlbum(context.Background(), nil, AlbumRequest{Album: album, Cache: cache}, nil)
	if err == nil {
		t.Fatalf("expandAlbum() error = nil, want an error for a nonexistent album directory")
	}
	if tasks != nil || state != nil {
		t.Errorf("expandAlbum() = (%v, %v), want (nil, nil) alongside the error", tasks, state)
	}
}

// TestExpandAlbumSkipsSubdirectories checks that a subdirectory sitting
// alongside a candidate media file never turns into a task of its own.
func TestExpandAlbumSkipsSubdirectories(t *testing.T) {
	albumDir := t.TempDir()
	copyFixtureFile(t, "photo/plain.jpg", filepath.Join(albumDir, "plain.jpg"))
	if err := os.Mkdir(filepath.Join(albumDir, "subdir"), 0o755); err != nil {
		t.Fatalf("mkdir subdir: %v", err)
	}

	album := &models.Album{Title: "with subdir", Path: albumDir}
	cache := scanner_cache.MakeAlbumCache()

	tasks, state, err := expandAlbum(context.Background(), nil, AlbumRequest{Album: album, Cache: cache}, nil)
	if err != nil {
		t.Fatalf("expandAlbum() error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expandAlbum() produced %d tasks, want 1 (subdirectory should be skipped)", len(tasks))
	}
	if want := filepath.Join(albumDir, "plain.jpg"); tasks[0].path != want {
		t.Errorf("task path = %q, want %q", tasks[0].path, want)
	}
	if state.total != 1 {
		t.Errorf("state.total = %d, want 1", state.total)
	}
}
