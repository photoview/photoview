package models

import (
	"database/sql"
)

func initializeSiteInfoRow(db *sql.DB) error {
	_, err := db.Exec("INSERT INTO site_info (initial_setup) VALUES (true)")
	if err != nil {
		return err
	}
	return nil
}

func GetSiteInfo(db *sql.DB) (*SiteInfo, error) {
	rows, err := db.Query("SELECT * FROM site_info")
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	var initialSetup bool
	var periodicScanInterval int
	var concurrentWorkers int

	if !rows.Next() {
		// Entry does not exist
		if err := initializeSiteInfoRow(db); err != nil {
			return nil, err
		}
		initialSetup = true
	} else {
		if err := rows.Scan(&initialSetup, &periodicScanInterval, &concurrentWorkers); err != nil {
			return nil, err
		}
	}

	return &SiteInfo{
		InitialSetup:         initialSetup,
		PeriodicScanInterval: periodicScanInterval,
		ConcurrentWorkers:    concurrentWorkers,
	}, nil
}
