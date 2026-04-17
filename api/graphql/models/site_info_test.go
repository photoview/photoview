package models_test

import (
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestSiteInfo(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	site_info, err := models.GetSiteInfo(db)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, models.DefaultSiteInfo(db), *site_info)

	site_info.InitialSetup = false
	site_info.PeriodicScanInterval = 360
	site_info.ConcurrentWorkers = 10

	if !assert.NoError(t, db.Session(&gorm.Session{AllowGlobalUpdate: true}).Save(&site_info).Error) {
		return
	}

	site_info, err = models.GetSiteInfo(db)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, models.SiteInfo{
		InitialSetup:         false,
		PeriodicScanInterval: 360,
		ConcurrentWorkers:    10,
		MapStyleLight:        nil,
		MapStyleDark:         nil,
	}, *site_info)

}

func TestSiteInfoMapStyles(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	site_info, err := models.GetSiteInfo(db)
	if !assert.NoError(t, err) {
		return
	}

	assert.Nil(t, site_info.MapStyleLight)
	assert.Nil(t, site_info.MapStyleDark)

	// Update map style URLs
	light := "https://example.com/light"
	dark := "https://example.com/dark"
	site_info.MapStyleLight = &light
	site_info.MapStyleDark = &dark

	if !assert.NoError(t, db.Session(&gorm.Session{AllowGlobalUpdate: true}).Save(&site_info).Error) {
		return
	}

	site_info, err = models.GetSiteInfo(db)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "https://example.com/light", *site_info.MapStyleLight)
	assert.Equal(t, "https://example.com/dark", *site_info.MapStyleDark)

	// Reset to nil
	site_info.MapStyleLight = nil
	site_info.MapStyleDark = nil

	if !assert.NoError(t, db.Session(&gorm.Session{AllowGlobalUpdate: true}).Save(&site_info).Error) {
		return
	}

	site_info, err = models.GetSiteInfo(db)
	if !assert.NoError(t, err) {
		return
	}

	assert.Nil(t, site_info.MapStyleLight)
	assert.Nil(t, site_info.MapStyleDark)
}
