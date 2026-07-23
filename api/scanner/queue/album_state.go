package queue

import (
	"context"
	"fmt"
	"os"
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/graphql/notification"
	"github.com/photoview/photoview/api/scanner/face_detection"
	"github.com/photoview/photoview/api/scanner/scanner_cache"
	"github.com/photoview/photoview/api/scanner/scanner_utils"
	"github.com/photoview/photoview/api/utils"
	"gorm.io/gorm"
)

// albumState is shared by every task produced for a single album scan.
// It is not tied to any particular batch of workers - a batch of workers is
// transient scheduling detail, this state is what carries meaning across the
// whole album's tasks.
type albumState struct {
	db       *gorm.DB
	album    *models.Album
	cache    *scanner_cache.AlbumScannerCache
	albumKey string
	total    int

	// skipAfter is set for a standalone single-media reprocess (queue.ProcessMedia):
	// it must not run album-wide cleanup, since allMediaIDs would only ever
	// contain the one media being reprocessed.
	skipAfter bool

	mu           sync.Mutex
	changedCount int
	remaining    int
	allMediaIDs  []int
	throttle     utils.Throttle
}

func newAlbumState(db *gorm.DB, album *models.Album, cache *scanner_cache.AlbumScannerCache, total int) *albumState {
	return &albumState{
		db:        db,
		album:     album,
		cache:     cache,
		albumKey:  utils.GenerateToken(),
		total:     total,
		remaining: total,
		throttle:  utils.NewThrottle(500 * time.Millisecond),
	}
}

// NotifyFoundNewMedia is called by a worker as soon as decide() confirms a
// task is for new media.
func (s *albumState) NotifyFoundNewMedia(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.throttle.Trigger(func() {
		notification.BroadcastNotification(&models.Notification{
			Key:     s.albumKey,
			Type:    models.NotificationTypeMessage,
			Header:  fmt.Sprintf("Found new media in album '%s'", s.album.Title),
			Content: fmt.Sprintf("Found %s", path),
		})
	})
}

// CompleteMedia is called by a worker exactly once per task, however it ended
// (skipped, errored, or processed). media is the task's media, if gather got
// far enough to determine/create it (nil otherwise) - its ID is recorded
// here so cleanup doesn't delete it, whatever happened afterwards. Only the
// ID is kept (not *models.Media) so this doesn't pin every Media object in
// memory for the whole scan. changed indicates whether this represents a
// real update (used only for the changedCount tally).
func (s *albumState) CompleteMedia(ctx context.Context, media *models.Media, changed bool) {
	s.mu.Lock()
	if media != nil {
		s.allMediaIDs = append(s.allMediaIDs, media.ID)
	}
	if changed {
		s.changedCount++
	}
	s.remaining--

	done := s.total - s.remaining
	total := s.total
	progress := float64(done) / float64(total) * 100.0
	s.throttle.Trigger(func() {
		notification.BroadcastNotification(&models.Notification{
			Key:      s.albumKey,
			Type:     models.NotificationTypeProgress,
			Header:   fmt.Sprintf("Processing media for album '%s'", s.album.Title),
			Content:  fmt.Sprintf("%d/%d processed", done, total),
			Progress: &progress,
		})
	})

	isLast := s.remaining == 0
	s.mu.Unlock()

	if isLast {
		s.completeAlbum(ctx)
	}
}

// completeAlbum runs once, when the last task for this album scan completes:
// delete stale media (files no longer found on disk) and broadcast a
// "scan complete" notification.
func (s *albumState) completeAlbum(ctx context.Context) {
	if s.skipAfter {
		return
	}

	if err := s.cleanupStaleMedia(ctx); err != nil {
		scanner_utils.ScannerError(ctx, "cleanup stale media for album (%s): %s", s.album.Path, err)
	}

	if s.changedCount > 0 {
		timeoutDelay := 2000
		notification.BroadcastNotification(&models.Notification{
			Key:      s.albumKey,
			Type:     models.NotificationTypeMessage,
			Positive: true,
			Header:   fmt.Sprintf("Done processing media for album '%s'", s.album.Title),
			Content:  "All media have been processed",
			Timeout:  &timeoutDelay,
		})
	}
}

// cleanupStaleMedia removes media rows (and their cache directories) that
// belong to this album but were not found during this scan.
func (s *albumState) cleanupStaleMedia(ctx context.Context) error {
	var staleMedia []models.Media

	query := s.db.Where("album_id = ?", s.album.ID)
	if len(s.allMediaIDs) > 0 {
		query = query.Where("NOT id IN (?)", s.allMediaIDs)
	}

	if err := query.Find(&staleMedia).Error; err != nil {
		return fmt.Errorf("get stale media from database: %w", err)
	}

	if len(staleMedia) == 0 {
		return nil
	}

	staleIDs := make([]int, len(staleMedia))
	for i, media := range staleMedia {
		staleIDs[i] = media.ID

		cachePath := path.Join(utils.MediaCachePath(), strconv.Itoa(s.album.ID), strconv.Itoa(media.ID))
		if err := os.RemoveAll(cachePath); err != nil {
			scanner_utils.ScannerError(ctx, "delete unused cache folder (%s): %s", cachePath, err)
		}
	}

	if err := s.db.Where("id IN (?)", staleIDs).Delete(&models.Media{}).Error; err != nil {
		return fmt.Errorf("delete stale media from database: %w", err)
	}

	if face_detection.GlobalFaceDetector != nil {
		if err := face_detection.GlobalFaceDetector.ReloadFacesFromDatabase(s.db); err != nil {
			return fmt.Errorf("reload faces from database: %w", err)
		}
	}

	return nil
}
