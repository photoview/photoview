package queue

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/face_detection"
	"github.com/photoview/photoview/api/scanner/scanner_cache"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/photoview/photoview/api/utils"
	"gorm.io/gorm"
)

// runFullPipeline drives mediaPath through the complete gather -> decide ->
// process -> persist pipeline via processMedia (the same path production
// code takes) and returns the task, so callers can inspect tsk.info.media
// and tsk.result afterwards. It fails the test if no media ended up
// persisted.
func runFullPipeline(t *testing.T, db *gorm.DB, album *models.Album, mediaPath string) *task {
	t.Helper()

	cache := scanner_cache.MakeAlbumCache()
	state := newAlbumState(db, album, cache, 1)
	tsk := newTask(album, cache, state, mediaPath, nil)

	w := newWorker(context.Background(), db)
	defer w.close()

	w.processMedia(tsk)

	if tsk.info.media == nil || tsk.info.media.ID == 0 {
		t.Fatalf("processMedia(%s) did not persist a media row", mediaPath)
	}
	return tsk
}

// TestProcessVideoFullPipeline drives a real video through the whole
// pipeline, covering processVideo/probeVideo end to end - both are 0%
// covered by the rest of this package's tests. It runs two fixtures: an
// already web-compatible mp4 (original is copied as-is, no transcode) and a
// non-web-compatible mov (must be transcoded into a web-playable mp4
// instead), since decide() only ever asks for one or the other.
func TestProcessVideoFullPipeline(t *testing.T) {
	test_utils.FilesystemTest(t)
	db := test_utils.DatabaseTest(t)

	if err := face_detection.InitializeFaceDetector(db); err != nil {
		t.Fatalf("initialize face detector: %v", err)
	}

	tests := []struct {
		name         string
		fixture      string
		wantOriginal bool
		wantWebVideo bool
	}{
		{"already web-compatible mp4: original copied, no transcode", "video/clip.mp4", true, false},
		{"non-web mov: transcoded into a web mp4 instead of copied", "video/clip.mov", false, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			albumDir := t.TempDir()
			mediaPath := filepath.Join(albumDir, "clip"+filepath.Ext(tc.fixture))
			copyFixtureFile(t, tc.fixture, mediaPath)

			album := &models.Album{Title: "video pipeline test", Path: albumDir}
			if err := db.Create(album).Error; err != nil {
				t.Fatalf("create album: %v", err)
			}

			tsk := runFullPipeline(t, db, album, mediaPath)

			if (tsk.result.original != nil) != tc.wantOriginal {
				t.Errorf("result.original present = %v, want %v", tsk.result.original != nil, tc.wantOriginal)
			}
			if (tsk.result.webVideo != nil) != tc.wantWebVideo {
				t.Errorf("result.webVideo present = %v, want %v", tsk.result.webVideo != nil, tc.wantWebVideo)
			}
			if tsk.result.videoThumb == nil {
				t.Errorf("result.videoThumb is nil, want a generated video thumbnail")
			}
			if tsk.result.blurhash == nil {
				t.Errorf("result.blurhash is nil, want a generated blurhash")
			}
			if tsk.result.videoMeta == nil {
				t.Errorf("result.videoMeta is nil, want probed video metadata")
			}

			var saved models.Media
			if err := db.Preload("VideoMetadata").Where("id = ?", tsk.info.media.ID).First(&saved).Error; err != nil {
				t.Fatalf("query persisted media: %v", err)
			}
			if saved.VideoMetadata == nil {
				t.Errorf("persisted media has no VideoMetadata, want it saved by persist()")
			}
		})
	}
}

// TestProcessPhotoReusesExistingURLNames checks that when a media is
// rescanned and only its cache file (not the database row) is missing,
// processPhoto reuses the existing MediaURL's name instead of generating a
// fresh one - the DB row for that purpose is updated in place, not
// duplicated.
func TestProcessPhotoReusesExistingURLNames(t *testing.T) {
	test_utils.FilesystemTest(t)
	db := test_utils.DatabaseTest(t)

	if err := face_detection.InitializeFaceDetector(db); err != nil {
		t.Fatalf("initialize face detector: %v", err)
	}

	albumDir := t.TempDir()
	mediaPath := filepath.Join(albumDir, "plain.jpg")
	copyFixtureFile(t, "photo/plain.jpg", mediaPath)

	album := &models.Album{Title: "reuse url names test", Path: albumDir}
	if err := db.Create(album).Error; err != nil {
		t.Fatalf("create album: %v", err)
	}

	first := runFullPipeline(t, db, album, mediaPath)
	media := first.info.media

	var thumbBefore models.MediaURL
	if err := db.Where("media_id = ? AND purpose = ?", media.ID, models.PhotoThumbnail).First(&thumbBefore).Error; err != nil {
		t.Fatalf("query thumbnail url after first pass: %v", err)
	}

	cachePath, err := utils.CachePathForMedia(album.ID, media.ID)
	if err != nil {
		t.Fatalf("cache path for media: %v", err)
	}
	thumbFile := filepath.Join(cachePath, thumbBefore.MediaName)
	if err := os.Remove(thumbFile); err != nil {
		t.Fatalf("remove cached thumbnail file: %v", err)
	}

	runFullPipeline(t, db, album, mediaPath)

	var thumbAfter models.MediaURL
	if err := db.Where("media_id = ? AND purpose = ?", media.ID, models.PhotoThumbnail).First(&thumbAfter).Error; err != nil {
		t.Fatalf("query thumbnail url after second pass: %v", err)
	}

	if thumbAfter.MediaName != thumbBefore.MediaName {
		t.Errorf("thumbnail MediaName changed across rescan: before=%q after=%q, want the same name reused", thumbBefore.MediaName, thumbAfter.MediaName)
	}
	if _, err := os.Stat(thumbFile); err != nil {
		t.Errorf("thumbnail file was not regenerated at the reused name: %v", err)
	}
}

// TestPersistSavesExifMetadata checks the result.exif != nil branch in
// persist(): a photo with real EXIF data must end up with an Exif
// association row in the database.
func TestPersistSavesExifMetadata(t *testing.T) {
	test_utils.FilesystemTest(t)
	db := test_utils.DatabaseTest(t)

	if err := face_detection.InitializeFaceDetector(db); err != nil {
		t.Fatalf("initialize face detector: %v", err)
	}

	albumDir := t.TempDir()
	mediaPath := filepath.Join(albumDir, "with_exif.jpg")
	copyFixtureFile(t, "photo/with_exif.jpg", mediaPath)

	album := &models.Album{Title: "exif persist test", Path: albumDir}
	if err := db.Create(album).Error; err != nil {
		t.Fatalf("create album: %v", err)
	}

	tsk := runFullPipeline(t, db, album, mediaPath)

	var saved models.Media
	if err := db.Preload("Exif").Where("id = ?", tsk.info.media.ID).First(&saved).Error; err != nil {
		t.Fatalf("query persisted media: %v", err)
	}
	if saved.Exif == nil {
		t.Fatalf("persisted media has no Exif row, want it saved by persist()")
	}
	if saved.Exif.Maker == nil || *saved.Exif.Maker != "Canon" {
		t.Errorf("Exif.Maker = %v, want %q", derefStr(saved.Exif.Maker), "Canon")
	}
	if saved.Exif.Camera == nil || *saved.Exif.Camera != "Canon EOS 500D" {
		t.Errorf("Exif.Camera = %v, want %q", derefStr(saved.Exif.Camera), "Canon EOS 500D")
	}
}

// TestPersistFinalizeCacheDirRenameFailure checks that persist() surfaces a
// clear error when the final os.Rename of a new media's cache directory
// fails, instead of silently losing the generated files. The rename is made
// to fail by deleting process()'s scratch directory out from under it right
// before persist() runs (e.g. simulating a concurrent cleanup) - simpler and
// more portable than pre-occupying the destination path, which would need
// to predict a not-yet-assigned database ID.
func TestPersistFinalizeCacheDirRenameFailure(t *testing.T) {
	test_utils.FilesystemTest(t)
	db := test_utils.DatabaseTest(t)

	if err := face_detection.InitializeFaceDetector(db); err != nil {
		t.Fatalf("initialize face detector: %v", err)
	}

	albumDir := t.TempDir()
	album := &models.Album{Title: "rename failure test", Path: albumDir}
	if err := db.Create(album).Error; err != nil {
		t.Fatalf("create album: %v", err)
	}

	targetPath := filepath.Join(albumDir, "target.jpg")
	copyFixtureFile(t, "photo/plain.jpg", targetPath)

	cache := scanner_cache.MakeAlbumCache()
	state := newAlbumState(db, album, cache, 1)
	tsk := newTask(album, cache, state, targetPath, nil)

	w := newWorker(context.Background(), db)
	defer w.close()

	info, err := w.gather(context.Background(), tsk)
	if err != nil {
		t.Fatalf("gather() error: %v", err)
	}
	tsk.info = info
	tsk.plan = decide(info)

	result, err := w.process(context.Background(), tsk)
	if err != nil {
		t.Fatalf("process() error: %v", err)
	}
	tsk.result = result

	if err := os.RemoveAll(result.pendingCachePath); err != nil {
		t.Fatalf("remove pending cache dir to force a rename failure: %v", err)
	}

	err = w.persist(context.Background(), tsk)
	if err == nil {
		t.Fatalf("persist() error = nil, want a finalize-cache-directory error")
	}
	if !strings.Contains(err.Error(), "finalize cache directory") {
		t.Errorf("persist() error = %q, want it to mention finalize cache directory", err.Error())
	}
}

// TestProcessPhotoDecodeFailureReturnsError checks that a file which passes
// type detection (extension + magic bytes say jpeg) but fails to actually
// decode makes process() return an error, instead of silently producing an
// empty/broken result.
func TestProcessPhotoDecodeFailureReturnsError(t *testing.T) {
	test_utils.FilesystemTest(t)
	db := test_utils.DatabaseTest(t)

	albumDir := t.TempDir()
	mediaPath := filepath.Join(albumDir, "undecodable.jpg")
	copyFixtureFile(t, "bad/undecodable.jpg", mediaPath)

	album := &models.Album{Title: "decode failure test", Path: albumDir}
	if err := db.Create(album).Error; err != nil {
		t.Fatalf("create album: %v", err)
	}

	cache := scanner_cache.MakeAlbumCache()
	state := newAlbumState(db, album, cache, 1)
	tsk := newTask(album, cache, state, mediaPath, nil)

	w := newWorker(context.Background(), db)
	defer w.close()

	info, err := w.gather(context.Background(), tsk)
	if err != nil {
		t.Fatalf("gather() error: %v", err)
	}
	if info.skip {
		t.Fatalf("info.skip = true, want the file to pass type detection and reach process()")
	}
	tsk.info = info
	tsk.plan = decide(info)

	if _, err := w.process(context.Background(), tsk); err == nil {
		t.Fatalf("process() error = nil, want an error decoding a corrupted jpeg")
	}
}
