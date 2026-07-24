package queue

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/scanner_cache"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/photoview/photoview/api/utils"
	"gorm.io/gorm"
)

// TestNewMediaNotVisibleBeforePersist drives a brand-new file through
// gather -> decide -> process by hand (instead of processMedia, so the
// intermediate state can be inspected) and checks that no Media row exists
// in the database until persist() actually runs - the whole point of
// deferring the insert is that nothing can observe a Media row with no
// thumbnail to go with it. It also checks that persist() ends up with the
// Media row, its MediaURL, and the cache file all in the final location,
// with the scratch pending directory gone.
func TestNewMediaNotVisibleBeforePersist(t *testing.T) {
	test_utils.FilesystemTest(t)
	db := test_utils.DatabaseTest(t)

	albumDir := t.TempDir()
	mediaPath := filepath.Join(albumDir, "photo.jpg")
	copyFixtureJPEG(t, mediaPath)

	album := &models.Album{Title: "deferred media test", Path: albumDir}
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
	if !info.isNewMedia {
		t.Fatalf("isNewMedia = false, want true for a never-before-seen path")
	}
	if info.media == nil || info.media.ID != 0 {
		t.Fatalf("info.media = %+v, want a non-nil media with ID == 0 before persist", info.media)
	}
	assertNoMediaRow(t, db, mediaPath)

	tsk.info = info
	tsk.plan = decide(info)

	result, err := w.process(context.Background(), tsk)
	if err != nil {
		t.Fatalf("process() error: %v", err)
	}
	if result.pendingCachePath == "" {
		t.Fatalf("result.pendingCachePath is empty, want a scratch directory for a new media")
	}
	if entries, err := os.ReadDir(result.pendingCachePath); err != nil || len(entries) == 0 {
		t.Fatalf("pending cache dir %s: ReadDir() = %v, %v, want at least one generated file", result.pendingCachePath, entries, err)
	}

	// Still nothing in the database - process() only touched the filesystem.
	assertNoMediaRow(t, db, mediaPath)

	tsk.result = result

	if err := w.persist(context.Background(), tsk); err != nil {
		t.Fatalf("persist() error: %v", err)
	}

	var media models.Media
	if err := db.Where("path_hash = ?", models.MD5Hash(mediaPath)).First(&media).Error; err != nil {
		t.Fatalf("expected a media row to exist after persist(), got error: %v", err)
	}
	if media.ID == 0 {
		t.Fatalf("persisted media has ID == 0")
	}

	var urlCount int64
	if err := db.Model(&models.MediaURL{}).Where("media_id = ?", media.ID).Count(&urlCount).Error; err != nil {
		t.Fatalf("count media urls: %v", err)
	}
	if urlCount == 0 {
		t.Fatalf("expected at least one media url after persist(), got 0")
	}

	finalPath := utils.MediaCacheLeafPath(album.ID, media.ID)
	if entries, err := os.ReadDir(finalPath); err != nil || len(entries) == 0 {
		t.Fatalf("final cache dir %s: ReadDir() = %v, %v, want the renamed generated files", finalPath, entries, err)
	}
	if _, err := os.Stat(result.pendingCachePath); !os.IsNotExist(err) {
		t.Fatalf("pending cache dir %s still exists after persist(), want it renamed away", result.pendingCachePath)
	}
}

// TestNewMediaPendingCacheCleanedUpOnPersistFailure checks that if persist()
// fails for a brand-new media, the scratch directory process() wrote its
// generated files to gets removed instead of leaking on disk forever. The
// album's ID is never actually saved to the database, so persist()'s
// UpsertMedia insert fails on the media.album_id foreign key - a reliable,
// direct way to make persist() fail after process() has already produced
// files, without needing to fake out the database connection itself.
func TestNewMediaPendingCacheCleanedUpOnPersistFailure(t *testing.T) {
	test_utils.FilesystemTest(t)
	db := test_utils.DatabaseTest(t)

	albumDir := t.TempDir()
	mediaPath := filepath.Join(albumDir, "photo.jpg")
	copyFixtureJPEG(t, mediaPath)

	album := &models.Album{Model: models.Model{ID: 999999}, Title: "never persisted", Path: albumDir}

	cache := scanner_cache.MakeAlbumCache()
	state := newAlbumState(db, album, cache, 1)
	tsk := newTask(album, cache, state, mediaPath, nil)

	w := newWorker(context.Background(), db)
	defer w.close()

	w.processMedia(tsk)

	if tsk.result.pendingCachePath == "" {
		t.Fatalf("tsk.result.pendingCachePath is empty, want process() to have used a scratch directory")
	}
	if _, err := os.Stat(tsk.result.pendingCachePath); !os.IsNotExist(err) {
		t.Fatalf("pending cache dir %s still exists after a failed persist(), want it cleaned up", tsk.result.pendingCachePath)
	}
	assertNoMediaRow(t, db, mediaPath)
}

func assertNoMediaRow(t *testing.T, db *gorm.DB, mediaPath string) {
	t.Helper()
	err := db.Where("path_hash = ?", models.MD5Hash(mediaPath)).First(&models.Media{}).Error
	if err == nil {
		t.Fatalf("expected no media row for %s yet, but one exists", mediaPath)
	}
	if err != gorm.ErrRecordNotFound {
		t.Fatalf("unexpected error checking for media row: %v", err)
	}
}

// copyFixtureJPEG copies a small, real, already-exercised test fixture image
// to dst, so the full process() pipeline (MIME detection, thumbnail
// generation, ...) has genuinely valid image bytes to work with.
func copyFixtureJPEG(t *testing.T, dst string) {
	t.Helper()
	src := test_utils.PathFromAPIRoot("scanner", "test_media", "orient", "up_arrow_90cw_web.jpg")
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("read fixture jpeg: %v", err)
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		t.Fatalf("write fixture jpeg copy: %v", err)
	}
}
