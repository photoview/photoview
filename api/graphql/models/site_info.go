package models

import (
	"database/sql"
)

func GetSiteInfo(db *sql.DB) (*SiteInfo, error) {
	rows, err := db.Query("SELECT * FROM site_info")
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	var initialSetup bool

	if !rows.Next() {
		// Entry does not exist
		_, err := db.Exec("INSERT INTO site_info (initial_setup) VALUES (true)")
		if err != nil {
			return nil, err
		}
		initialSetup = true
	} else {
		if err := rows.Scan(&initialSetup); err != nil {
			return nil, err
		}
	}

	return &SiteInfo{
		InitialSetup: initialSetup,
	}, nil
}
