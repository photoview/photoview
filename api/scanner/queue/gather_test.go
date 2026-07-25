package queue

import (
	"context"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/scanner_cache"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/photoview/photoview/api/utils"
	"gorm.io/gorm"
)

// TestGatherSkipsIgnoredFile checks the .photoviewignore skip branch: a file
// matching an album's ignore lines must be skipped, without ever reaching
// findOrBuildMedia/the database.
func TestGatherSkipsIgnoredFile(t *testing.T) {
	test_utils.FilesystemTest(t)
	db := test_utils.DatabaseTest(t)

	albumDir := t.TempDir()
	mediaPath := filepath.Join(albumDir, "plain.jpg")
	copyFixtureFile(t, "photo/plain.jpg", mediaPath)

	album := &models.Album{Title: "ignore test", Path: albumDir}
	if err := db.Create(album).Error; err != nil {
		t.Fatalf("create album: %v", err)
	}

	cache := scanner_cache.MakeAlbumCache()
	cache.InsertAlbumIgnore(albumDir, []string{"*.jpg"})

	info := gatherForPath(t, db, album, cache, mediaPath)
	if !info.skip {
		t.Errorf("info.skip = false, want true for a file matching the album's ignore lines")
	}
	assertNoMediaRow(t, db, mediaPath)
}

// TestGatherSkipsUnsupportedType checks that a file with an unsupported
// media type is skipped before any database/filesystem lookups happen.
func TestGatherSkipsUnsupportedType(t *testing.T) {
	test_utils.FilesystemTest(t)
	db := test_utils.DatabaseTest(t)

	albumDir := t.TempDir()
	mediaPath := filepath.Join(albumDir, "notes.pdf")
	copyFixtureFile(t, "unsupported/notes.pdf", mediaPath)

	album := &models.Album{Title: "unsupported type test", Path: albumDir}
	if err := db.Create(album).Error; err != nil {
		t.Fatalf("create album: %v", err)
	}

	info := gatherForPath(t, db, album, scanner_cache.MakeAlbumCache(), mediaPath)
	if !info.skip {
		t.Errorf("info.skip = false, want true for an unsupported file type")
	}
	assertNoMediaRow(t, db, mediaPath)
}

// TestGatherRawWebMutualExclusion covers the two skip checks that keep a raw
// file and its web-compatible derivative from both being scanned: with raw
// processing disabled, the raw file itself is skipped; with it enabled
// (default), the web derivative is skipped instead since the raw file will
// be processed in its place.
func TestGatherRawWebMutualExclusion(t *testing.T) {
	test_utils.FilesystemTest(t)
	db := test_utils.DatabaseTest(t)

	tests := []struct {
		name        string
		disableRaw  bool
		fixtureFile string
		wantSkip    bool
	}{
		{"raw processing disabled: non-web raw file is skipped", true, "photo.tiff", true},
		{"raw processing disabled: web file is unaffected", true, "photo.jpg", false},
		{"raw processing enabled: web file with a raw counterpart is skipped", false, "photo.jpg", true},
		{"raw processing enabled: the raw file itself is not skipped by this check", false, "photo.tiff", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(utils.EnvDisableRawProcessing.GetName(), strconv.FormatBool(tc.disableRaw))

			albumDir := t.TempDir()
			copyFixtureDir(t, "raw_pair", albumDir)
			mediaPath := filepath.Join(albumDir, tc.fixtureFile)

			album := &models.Album{Title: "raw/web mutual exclusion test", Path: albumDir}
			if err := db.Create(album).Error; err != nil {
				t.Fatalf("create album: %v", err)
			}

			info := gatherForPath(t, db, album, scanner_cache.MakeAlbumCache(), mediaPath)
			if info.skip != tc.wantSkip {
				t.Errorf("info.skip = %v, want %v", info.skip, tc.wantSkip)
			}
		})
	}
}

// TestGatherDetectsSidecarAndHash checks that gather() finds a raw file's
// .xmp sidecar and records its hash - the only place hashFile is actually
// called from production code.
func TestGatherDetectsSidecarAndHash(t *testing.T) {
	test_utils.FilesystemTest(t)
	db := test_utils.DatabaseTest(t)
	t.Setenv(utils.EnvDisableRawProcessing.GetName(), "false")

	albumDir := t.TempDir()
	copyFixtureDir(t, "raw_pair", albumDir)
	mediaPath := filepath.Join(albumDir, "photo.tiff")
	sidecarPath := mediaPath + ".xmp"

	album := &models.Album{Title: "sidecar test", Path: albumDir}
	if err := db.Create(album).Error; err != nil {
		t.Fatalf("create album: %v", err)
	}

	info := gatherForPath(t, db, album, scanner_cache.MakeAlbumCache(), mediaPath)

	if info.sidecarPath == nil || *info.sidecarPath != sidecarPath {
		t.Fatalf("info.sidecarPath = %v, want %q", info.sidecarPath, sidecarPath)
	}

	wantHash := hashFile(context.Background(), sidecarPath)
	if wantHash == nil {
		t.Fatalf("hashFile() on the sidecar file itself returned nil")
	}
	if info.sidecarHash == nil || *info.sidecarHash != *wantHash {
		t.Errorf("info.sidecarHash = %v, want %v", info.sidecarHash, *wantHash)
	}
}

// TestGatherDetectsWebCounterpart checks that a raw file with a
// web-compatible derivative sitting next to it gets hasCounterpart/
// webCounterpart set.
func TestGatherDetectsWebCounterpart(t *testing.T) {
	test_utils.FilesystemTest(t)
	db := test_utils.DatabaseTest(t)
	t.Setenv(utils.EnvDisableRawProcessing.GetName(), "false")

	albumDir := t.TempDir()
	copyFixtureDir(t, "raw_pair", albumDir)
	mediaPath := filepath.Join(albumDir, "photo.tiff")

	album := &models.Album{Title: "web counterpart test", Path: albumDir}
	if err := db.Create(album).Error; err != nil {
		t.Fatalf("create album: %v", err)
	}

	info := gatherForPath(t, db, album, scanner_cache.MakeAlbumCache(), mediaPath)

	if !info.hasCounterpart {
		t.Fatalf("info.hasCounterpart = false, want true")
	}
	if info.webCounterpart == "" {
		t.Errorf("info.webCounterpart is empty, want the path to photo.jpg")
	}
}

// TestGatherVideoMissingURLDetection checks the VideoWeb/VideoThumbnail
// missing-cache-file detection for an already-known video, which is
// otherwise indistinguishable in coverage from the photo purposes.
func TestGatherVideoMissingURLDetection(t *testing.T) {
	test_utils.FilesystemTest(t)
	db := test_utils.DatabaseTest(t)

	albumDir := t.TempDir()
	mediaPath := filepath.Join(albumDir, "clip.mp4")
	copyFixtureFile(t, "video/clip.mp4", mediaPath)

	album := &models.Album{Title: "video missing url test", Path: albumDir}
	if err := db.Create(album).Error; err != nil {
		t.Fatalf("create album: %v", err)
	}

	media := &models.Media{
		Title:    "clip.mp4",
		Path:     mediaPath,
		AlbumID:  album.ID,
		Type:     models.MediaTypeVideo,
		DateShot: time.Now(),
	}
	if err := db.Create(media).Error; err != nil {
		t.Fatalf("create media: %v", err)
	}

	for _, purpose := range []models.MediaPurpose{models.VideoWeb, models.VideoThumbnail} {
		url := &models.MediaURL{
			MediaID:     media.ID,
			MediaName:   "does-not-exist.mp4",
			Purpose:     purpose,
			ContentType: "video/mp4",
		}
		if err := db.Create(url).Error; err != nil {
			t.Fatalf("create media url (%s): %v", purpose, err)
		}
	}

	info := gatherForPath(t, db, album, scanner_cache.MakeAlbumCache(), mediaPath)

	if info.isNewMedia {
		t.Fatalf("info.isNewMedia = true, want false (media row already existed)")
	}
	if !info.webVideoMissing {
		t.Errorf("info.webVideoMissing = false, want true")
	}
	if !info.videoThumbMissing {
		t.Errorf("info.videoThumbMissing = false, want true")
	}
}

// gatherForPath builds a minimal task for mediaPath and runs gather() on it,
// failing the test immediately on error. It builds a bare worker directly
// rather than going through newWorker, since gather() only ever touches
// w.db - going through newWorker would spawn a real, unused exiftool
// subprocess on every call.
func gatherForPath(t *testing.T, db *gorm.DB, album *models.Album, cache *scanner_cache.AlbumScannerCache, mediaPath string) gatheredInfo {
	t.Helper()

	state := newAlbumState(db, album, cache, 1)
	tsk := newTask(album, cache, state, mediaPath, nil)

	w := &worker{ctx: context.Background(), db: db}

	info, err := w.gather(context.Background(), tsk)
	if err != nil {
		t.Fatalf("gather() error: %v", err)
	}
	return info
}
