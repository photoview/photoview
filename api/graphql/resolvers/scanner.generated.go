package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.55

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/photoview/photoview/api/database/drivers"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/periodic_scanner"
	"github.com/photoview/photoview/api/scanner/scanner_queue"
	"gorm.io/gorm"
)

// ScanAll is the resolver for the scanAll field.
func (r *mutationResolver) ScanAll(ctx context.Context) (*models.ScannerResult, error) {
	err := scanner_queue.AddAllToQueue()
	if err != nil {
		return nil, err
	}

	startMessage := "Scanner started"

	return &models.ScannerResult{
		Finished: false,
		Success:  true,
		Message:  &startMessage,
	}, nil
}

// ScanUser is the resolver for the scanUser field.
func (r *mutationResolver) ScanUser(ctx context.Context, userID int) (*models.ScannerResult, error) {
	var user models.User
	if err := r.DB(ctx).First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("get user from database: %w", err)
	}

	scanner_queue.AddUserToQueue(&user)

	startMessage := "Scanner started"
	return &models.ScannerResult{
		Finished: false,
		Success:  true,
		Message:  &startMessage,
	}, nil
}

// SetPeriodicScanInterval is the resolver for the setPeriodicScanInterval field.
func (r *mutationResolver) SetPeriodicScanInterval(ctx context.Context, interval int) (int, error) {
	db := r.DB(ctx)
	if interval < 0 {
		return 0, errors.New("interval must be 0 or above")
	}

	if err := db.
		Session(&gorm.Session{AllowGlobalUpdate: true}).
		Model(&models.SiteInfo{}).
		Update("periodic_scan_interval", interval).
		Error; err != nil {

		return 0, err
	}

	var siteInfo models.SiteInfo
	if err := db.First(&siteInfo).Error; err != nil {
		return 0, err
	}

	periodic_scanner.ChangePeriodicScanInterval(time.Duration(siteInfo.PeriodicScanInterval) * time.Second)

	return siteInfo.PeriodicScanInterval, nil
}

// SetScannerConcurrentWorkers is the resolver for the setScannerConcurrentWorkers field.
func (r *mutationResolver) SetScannerConcurrentWorkers(ctx context.Context, workers int) (int, error) {
	db := r.DB(ctx)
	if workers < 1 {
		return 0, errors.New("concurrent workers must at least be 1")
	}

	if workers > 1 && drivers.DatabaseDriverFromEnv() == drivers.SQLITE {
		return 0, errors.New("multiple workers not supported for SQLite databases")
	}

	if err := db.
		Session(&gorm.Session{AllowGlobalUpdate: true}).
		Model(&models.SiteInfo{}).
		Update("concurrent_workers", workers).
		Error; err != nil {

		return 0, err
	}

	var siteInfo models.SiteInfo
	if err := db.First(&siteInfo).Error; err != nil {
		return 0, err
	}

	scanner_queue.ChangeScannerConcurrentWorkers(siteInfo.ConcurrentWorkers)

	return siteInfo.ConcurrentWorkers, nil
}