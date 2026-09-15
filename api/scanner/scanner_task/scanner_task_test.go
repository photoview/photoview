package scanner_task

import (
	"context"
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/scanner_cache"
	"github.com/photoview/photoview/api/test_utils"
)

func TestWithoutCancelOutlivesTheJobAndKeepsItsValues(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	jobCtx, cancel := context.WithCancel(context.Background())
	album := &models.Album{Title: "album"}
	cache := scanner_cache.MakeAlbumCache()

	ctx := NewTaskContext(jobCtx, db, album, cache)
	detached := ctx.WithoutCancel()

	cancel()

	if ctx.Err() == nil {
		t.Fatal("expected the job context to be cancelled")
	}

	if err := detached.Err(); err != nil {
		t.Errorf("the detached context was cancelled with the job: %v", err)
	}

	if detached.GetAlbum() != album || detached.GetCache() != cache {
		t.Error("the detached context lost its album or cache")
	}

	// The database handle stored in the job context carries the job's
	// cancellation into every query, so it has to be re-bound, not copied.
	if err := ctx.GetDB().Statement.Context.Err(); err == nil {
		t.Fatal("expected the job context's database handle to be cancelled")
	}

	if err := detached.GetDB().Statement.Context.Err(); err != nil {
		t.Errorf("the detached database handle still carries the job's cancellation: %v", err)
	}

	var count int64
	if err := detached.GetDB().Model(&models.Album{}).Count(&count).Error; err != nil {
		t.Errorf("a query through the detached handle failed after the job was cancelled: %v", err)
	}
}
