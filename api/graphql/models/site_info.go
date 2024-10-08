package models

import (
	"fmt"

	db_drivers "github.com/photoview/photoview/api/database/drivers"
	"gorm.io/gorm"
)

type SiteInfo struct {
	InitialSetup         bool `gorm:"not null"`
	PeriodicScanInterval int  `gorm:"not null"`
	ConcurrentWorkers    int  `gorm:"not null"`
	ThumbnailMethod   	 ThumbnailFilter  `gorm:"not null"`
}

func (SiteInfo) TableName() string {
	return "site_info"
}

func DefaultSiteInfo(db *gorm.DB) SiteInfo {
	defaultConcurrentWorkers := 3
	if db_drivers.SQLITE.MatchDatabase(db) {
		defaultConcurrentWorkers = 1
	}

	return SiteInfo{
		InitialSetup:         true,
		PeriodicScanInterval: 0,
		ConcurrentWorkers:    defaultConcurrentWorkers,
		ThumbnailMethod:			ThumbnailFilterNearestNeighbor,
	}
}

// GetSiteInfo gets the site info row from the database, and creates it if it does not exist
func GetSiteInfo(db *gorm.DB) (*SiteInfo, error) {

	var siteInfo []*SiteInfo

	if err := db.Limit(1).Find(&siteInfo).Error; err != nil {
		return nil, fmt.Errorf("get site info from database: %w", err)
	}

	if len(siteInfo) == 0 {
		newSiteInfo := DefaultSiteInfo(db)

		if err := db.Create(&newSiteInfo).Error; err != nil {
			return nil, fmt.Errorf("initialize site_info: %w", err)
		}

		return &newSiteInfo, nil
	} else {
		return siteInfo[0], nil
	}
}
