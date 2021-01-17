package models

import (
	db_drivers "github.com/photoview/photoview/api/database/drivers"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type SiteInfo struct {
	InitialSetup         bool `gorm:"not null"`
	PeriodicScanInterval int  `gorm:"not null"`
	ConcurrentWorkers    int  `gorm:"not null"`
}

func (SiteInfo) TableName() string {
	return "site_info"
}

// GetSiteInfo gets the site info row from the database, and creates it if it does not exist
func GetSiteInfo(db *gorm.DB) (*SiteInfo, error) {

	var siteInfo SiteInfo

	if err := db.First(&siteInfo).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {

			defaultConcurrentWorkers := 3
			if db_drivers.DatabaseDriver() == db_drivers.DatabaseDriverSqlite {
				defaultConcurrentWorkers = 1
			}

			siteInfo = SiteInfo{
				InitialSetup:         true,
				PeriodicScanInterval: 0,
				ConcurrentWorkers:    defaultConcurrentWorkers,
			}

			if err := db.Create(&siteInfo).Error; err != nil {
				return nil, errors.Wrap(err, "initialize site_info")
			}
		} else {
			return nil, errors.Wrap(err, "get site info from database")
		}
	}

	return &siteInfo, nil
}
