package test_utils

import (
	"fmt"
	"github.com/photoview/photoview/api/database/drivers"
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

func ScannerCleanup(t *testing.T, db *gorm.DB) {
	releventModels := []interface{}{
		&models.User{},
		&models.Media{},
		&models.MediaURL{},
		&models.Album{},
		&models.FaceGroup{},
		&models.ImageFace{},
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		db_driver := drivers.DatabaseDriverFromEnv()

		if db_driver == drivers.MYSQL {
			if err := tx.Exec("SET FOREIGN_KEY_CHECKS = 0;").Error; err != nil {
				return err
			}
		}
		dry_run := tx.Session(&gorm.Session{DryRun: true})
		for _, model := range releventModels {
			// get table name of model structure
			table := dry_run.Find(model).Statement.Table

			switch db_driver {
			case drivers.POSTGRES:
				if err := tx.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)).Error; err != nil {
					return err
				}
			case drivers.MYSQL:
				if err := tx.Exec(fmt.Sprintf("TRUNCATE TABLE %s", table)).Error; err != nil {
					return err
				}
			case drivers.SQLITE:
				if err := tx.Exec(fmt.Sprintf("DELETE FROM %s", table)).Error; err != nil {
					return err
				}
			}

		}

		if db_driver == drivers.MYSQL {
			if err := tx.Exec("SET FOREIGN_KEY_CHECKS = 1;").Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Error(err)
	}
}
