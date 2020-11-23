package models

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type SiteInfo struct {
	gorm.Model
	InitialSetup         bool
	PeriodicScanInterval int
	ConcurrentWorkers    int
}

// GetSiteInfo gets the site info row from the database, and creates it if it does not exist
func GetSiteInfo(db *gorm.DB) (*SiteInfo, error) {

	var siteInfo SiteInfo

	err := db.FirstOrCreate(&siteInfo, SiteInfo{
		InitialSetup:         true,
		PeriodicScanInterval: 0,
		ConcurrentWorkers:    3,
	}).Error
	if err != nil {
		return nil, errors.Wrap(err, "get site info from database")
	}

	return &siteInfo, nil
}
