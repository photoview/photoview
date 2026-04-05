package resolvers

import (
	"github.com/photoview/photoview/api/graphql/models"
	"gorm.io/gorm"
)

func updateSiteInfoField(db *gorm.DB, column string, value string) (string, error) {
	if err := db.
		Session(&gorm.Session{AllowGlobalUpdate: true}).
		Model(&models.SiteInfo{}).
		Update(column, value).
		Error; err != nil {
		return "", err
	}

	return value, nil
}
