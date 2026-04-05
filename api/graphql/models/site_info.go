package models

import (
	db_drivers "github.com/photoview/photoview/api/database/drivers"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const (
	// DefaultMapStyleLight is the default map tile style URL for light mode.
	// Keep in sync with struct tag defaults below and ui/src/components/maplibre/useMapStyles.ts
	DefaultMapStyleLight = "https://tiles.openfreemap.org/styles/positron"
	// DefaultMapStyleDark is the default map tile style URL for dark mode.
	// Keep in sync with struct tag defaults below and ui/src/components/maplibre/useMapStyles.ts
	DefaultMapStyleDark = "https://tiles.openfreemap.org/styles/dark"
)

type SiteInfo struct {
	InitialSetup         bool   `gorm:"not null"`
	PeriodicScanInterval int    `gorm:"not null"`
	ConcurrentWorkers    int    `gorm:"not null"`
	// Struct tag default must be a string literal; keep in sync with DefaultMapStyleLight
	MapStyleLight string `gorm:"not null;default:https://tiles.openfreemap.org/styles/positron"`
	// Struct tag default must be a string literal; keep in sync with DefaultMapStyleDark
	MapStyleDark string `gorm:"not null;default:https://tiles.openfreemap.org/styles/dark"`
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
		MapStyleLight:        DefaultMapStyleLight,
		MapStyleDark:         DefaultMapStyleDark,
	}
}

// GetSiteInfo gets the site info row from the database, and creates it if it does not exist
func GetSiteInfo(db *gorm.DB) (*SiteInfo, error) {

	var siteInfo []*SiteInfo

	if err := db.Limit(1).Find(&siteInfo).Error; err != nil {
		return nil, errors.Wrap(err, "get site info from database")
	}

	if len(siteInfo) == 0 {
		newSiteInfo := DefaultSiteInfo(db)

		if err := db.Create(&newSiteInfo).Error; err != nil {
			return nil, errors.Wrap(err, "initialize site_info")
		}

		return &newSiteInfo, nil
	} else {
		return siteInfo[0], nil
	}
}
