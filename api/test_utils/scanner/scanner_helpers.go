package scanner_utils

import (
	"testing"
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/scanner_queue"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func RunScannerOnUser(t *testing.T, db *gorm.DB, fs afero.Fs, cacheFs afero.Fs, user *models.User) {
	start := time.Now()
	defer func() {
		dur := time.Now().Sub(start)
		t.Logf("RunScannerOnUser(user(id:%d)) took %s.", user.ID, dur)
	}()

	if !assert.NoError(t, scanner_queue.InitializeScannerQueue(db, fs, cacheFs)) {
		return
	}

	if !assert.NoError(t, scanner_queue.AddUserToQueue(user)) {
		return
	}

	// wait for all jobs to finish
	scanner_queue.CloseScannerQueue()
}

func RunScannerAll(t *testing.T, db *gorm.DB, fs afero.Fs, cacheFs afero.Fs) {
	start := time.Now()
	defer func() {
		dur := time.Now().Sub(start)
		t.Logf("RunScannerAll() took %s.", dur)
	}()

	if !assert.NoError(t, scanner_queue.InitializeScannerQueue(db, fs, cacheFs)) {
		return
	}

	if !assert.NoError(t, scanner_queue.AddAllToQueue()) {
		return
	}

	// wait for all jobs to finish
	scanner_queue.CloseScannerQueue()
}
