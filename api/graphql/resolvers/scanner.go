package resolvers

import (
	"context"
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func (r *mutationResolver) ScanAll(ctx context.Context) (*models.ScannerResult, error) {
	err := scanner.AddAllToQueue()
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

func (r *mutationResolver) ScanUser(ctx context.Context, userID int) (*models.ScannerResult, error) {

	var user models.User
	if err := r.Database.First(&user, userID).Error; err != nil {
		return nil, errors.Wrap(err, "get user from database")
	}

	scanner.AddUserToQueue(&user)

	startMessage := "Scanner started"
	return &models.ScannerResult{
		Finished: false,
		Success:  true,
		Message:  &startMessage,
	}, nil
}

func (r *mutationResolver) SetPeriodicScanInterval(ctx context.Context, interval int) (int, error) {
	if interval < 0 {
		return 0, errors.New("interval must be 0 or above")
	}

	if err := r.Database.Session(&gorm.Session{AllowGlobalUpdate: true}).Model(&models.SiteInfo{}).Update("periodic_scan_interval", interval).Error; err != nil {
		return 0, err
	}

	var siteInfo models.SiteInfo
	if err := r.Database.First(&siteInfo).Error; err != nil {
		return 0, err
	}

	scanner.ChangePeriodicScanInterval(time.Duration(siteInfo.PeriodicScanInterval) * time.Second)

	return siteInfo.PeriodicScanInterval, nil
}

func (r *mutationResolver) SetScannerConcurrentWorkers(ctx context.Context, workers int) (int, error) {
	if workers < 1 {
		return 0, errors.New("concurrent workers must at least be 1")
	}

	if err := r.Database.Session(&gorm.Session{AllowGlobalUpdate: true}).Model(&models.SiteInfo{}).Update("concurrent_workers", workers).Error; err != nil {
		return 0, err
	}

	var siteInfo models.SiteInfo
	if err := r.Database.First(&siteInfo).Error; err != nil {
		return 0, err
	}

	scanner.ChangeScannerConcurrentWorkers(siteInfo.ConcurrentWorkers)

	return siteInfo.ConcurrentWorkers, nil
}
