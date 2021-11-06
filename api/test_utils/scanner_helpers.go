package test_utils

import (
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func RunScannerOnUser(t *testing.T, db *gorm.DB, user *models.User) {
	if !assert.NoError(t, scanner.InitializeScannerQueue(db)) {
		return
	}

	if !assert.NoError(t, scanner.AddUserToQueue(user)) {
		return
	}

	// wait for all jobs to finish
	scanner.CloseScannerQueue()
}

func RunScannerAll(t *testing.T, db *gorm.DB) {
	if !assert.NoError(t, scanner.InitializeScannerQueue(db)) {
		return
	}

	if !assert.NoError(t, scanner.AddAllToQueue()) {
		return
	}

	// wait for all jobs to finish
	scanner.CloseScannerQueue()
}
