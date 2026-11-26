package scanner_utils

import (
	"context"
	"testing"
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner"
	"github.com/photoview/photoview/api/scanner/queue"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func RunScannerOnUser(t *testing.T, db *gorm.DB, user *models.User) {
	start := time.Now()
	defer func() {
		dur := time.Since(start)
		t.Logf("RunScannerOnUser(user(id:%d)) took %s.", user.ID, dur)
	}()

	if !assert.NoError(t, queue.Initialize(context.Background(), db)) {
		return
	}
	defer queue.Close()

	if !assert.NoError(t, scanner.AddUser(db, user)) {
		return
	}
}

func RunScannerAll(t *testing.T, db *gorm.DB) {
	start := time.Now()
	defer func() {
		dur := time.Since(start)
		t.Logf("RunScannerAll() took %s.", dur)
	}()

	if !assert.NoError(t, queue.Initialize(context.Background(), db)) {
		return
	}
	defer queue.Close()

	if !assert.NoError(t, scanner.AddAll(db)) {
		return
	}
}
