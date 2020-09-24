package models

import (
	"database/sql"

	"github.com/pkg/errors"
)

func initializeSiteInfoRow(db *sql.DB) (*SiteInfo, error) {
	_, err := db.Exec("INSERT INTO site_info (initial_setup, periodic_scan_interval, concurrent_workers) VALUES (true, 0, 3)")
	if err != nil {
		return nil, errors.Wrap(err, "initialize site_info row")
	}

	siteInfo := &SiteInfo{}

	row := db.QueryRow("SELECT * FROM site_info")
	if err := row.Scan(&siteInfo.InitialSetup, &siteInfo.PeriodicScanInterval, &siteInfo.ConcurrentWorkers); err != nil {
		return nil, errors.Wrap(err, "get site_info row after initialization")
	}

	return siteInfo, nil
}

// GetSiteInfo gets the site info row from the database, and creates it if it does not exist
func GetSiteInfo(db *sql.DB) (*SiteInfo, error) {
	rows, err := db.Query("SELECT * FROM site_info")
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	siteInfo := &SiteInfo{}

	if !rows.Next() {
		// Entry does not exist
		siteInfo, err = initializeSiteInfoRow(db)
		if err != nil {
			return nil, err
		}
	} else {
		if err := rows.Scan(&siteInfo.InitialSetup, &siteInfo.PeriodicScanInterval, &siteInfo.ConcurrentWorkers); err != nil {
			return nil, err
		}
	}

	return siteInfo, nil
}
