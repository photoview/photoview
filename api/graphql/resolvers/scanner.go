package resolvers

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/viktorstrate/photoview/api/graphql/models"
	"github.com/viktorstrate/photoview/api/scanner"
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
	row := r.Database.QueryRow("SELECT * FROM user WHERE user_id = ?", userID)
	user, err := models.NewUserFromRow(row)
	if err != nil {
		return nil, errors.Wrap(err, "get user from database")
	}

	scanner.AddUserToQueue(user)

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

	_, err := r.Database.Exec("UPDATE site_info SET periodic_scan_interval = ?", interval)
	if err != nil {
		return 0, err
	}

	var dbInterval int

	row := r.Database.QueryRow("SELECT periodic_scan_interval FROM site_info")
	if err = row.Scan(&dbInterval); err != nil {
		return 0, err
	}

	scanner.ChangePeriodicScanInterval(time.Duration(dbInterval) * time.Second)

	return dbInterval, nil
}

func (r *mutationResolver) SetScannerConcurrentWorkers(ctx context.Context, workers int) (int, error) {
	if workers < 0 {
		return 0, errors.New("concurrent workers must be positive")
	}

	_, err := r.Database.Exec("UPDATE site_info SET concurrent_workers = ?", workers)
	if err != nil {
		return 0, err
	}

	var dbWorkers int

	row := r.Database.QueryRow("SELECT concurrent_workers FROM site_info")
	if err = row.Scan(&dbWorkers); err != nil {
		return 0, err
	}

	scanner.ChangeScannerConcurrentWorkers(dbWorkers)

	return dbWorkers, nil
}
