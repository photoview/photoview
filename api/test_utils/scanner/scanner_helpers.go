package scanner_utils

import (
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/queue"
	"gorm.io/gorm"
)

func RunScannerOnUser(t *testing.T, db *gorm.DB, user *models.User) {
	queue, err := queue.NewQueue(db)
	if err != nil {
		t.Fatalf("create queue error: %v", err)
		return
	}
	defer queue.Close()

	if err := queue.AddUserAlbums(t.Context(), user); err != nil {
		t.Fatalf("scan all albums error: %v", err)
		return
	}

	queue.ConsumeAllBacklog(t.Context())
}

func RunScannerAll(t *testing.T, db *gorm.DB) {
	queue, err := queue.NewQueue(db)
	if err != nil {
		t.Fatalf("create queue error: %v", err)
		return
	}
	defer queue.Close()

	if err := queue.AddAllAlbums(t.Context()); err != nil {
		t.Fatalf("scan all albums error: %v", err)
		return
	}

	queue.ConsumeAllBacklog(t.Context())
}
