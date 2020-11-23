package models

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// func initializeSiteInfoRow(db *gorm.DB) (*SiteInfo, error) {
// 	_, err := db.Exec("INSERT INTO site_info (initial_setup, periodic_scan_interval, concurrent_workers) VALUES (true, 0, 3)")
// 	if err != nil {
// 		return nil, errors.Wrap(err, "initialize site_info row")
// 	}

// 	siteInfo := &SiteInfo{}

// 	row := db.QueryRow("SELECT * FROM site_info")
// 	if err := row.Scan(&siteInfo.InitialSetup, &siteInfo.PeriodicScanInterval, &siteInfo.ConcurrentWorkers); err != nil {
// 		return nil, errors.Wrap(err, "get site_info row after initialization")
// 	}

// 	return siteInfo, nil
// }

// GetSiteInfo gets the site info row from the database, and creates it if it does not exist
func GetSiteInfo(db *gorm.DB) (*SiteInfo, error) {

	var siteInfo SiteInfo

	result := db.FirstOrCreate(&siteInfo, SiteInfo{
		InitialSetup:         true,
		PeriodicScanInterval: 0,
		ConcurrentWorkers:    3,
	})
	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "get site info from database")
	}

	// rows, err := db.Query("SELECT * FROM site_info")
	// defer rows.Close()
	// if err != nil {
	// 	return nil, err
	// }

	// siteInfo := &SiteInfo{}

	// if !rows.Next() {
	// 	// Entry does not exist
	// 	siteInfo, err = initializeSiteInfoRow(db)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// } else {
	// 	if err := rows.Scan(&siteInfo.InitialSetup, &siteInfo.PeriodicScanInterval, &siteInfo.ConcurrentWorkers); err != nil {
	// 		return nil, err
	// 	}
	// }

	return &siteInfo, nil
}
