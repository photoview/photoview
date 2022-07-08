package test_utils

import (
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/scanner_queue"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func RunScannerOnUser(t *testing.T, db *gorm.DB, user *models.User) {
	if !assert.NoError(t, scanner_queue.InitializeScannerQueue(db)) {
		return
	}

	if !assert.NoError(t, scanner_queue.AddUserToQueue(user)) {
		return
	}

	// wait for all jobs to finish
	scanner_queue.CloseScannerQueue()
}

func RunScannerAll(t *testing.T, db *gorm.DB) {
	if !assert.NoError(t, scanner_queue.InitializeScannerQueue(db)) {
		return
	}

	if !assert.NoError(t, scanner_queue.AddAllToQueue()) {
		return
	}

	// wait for all jobs to finish
	scanner_queue.CloseScannerQueue()
}
