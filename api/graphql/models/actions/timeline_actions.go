package actions

import (
	"time"

	"github.com/photoview/photoview/api/database/drivers"
	"github.com/photoview/photoview/api/graphql/models"
	"gorm.io/gorm"
)

func MyTimeline(db *gorm.DB, user *models.User, paginate *models.Pagination, onlyFavorites *bool, fromDate *time.Time) ([]*models.Media, error) {

	query := db.
		Joins("JOIN albums ON media.album_id = albums.id").
		Where("albums.id IN (?)", db.Table("user_albums").Select("user_albums.album_id").Where("user_id = ?", user.ID))

	switch drivers.GetDatabaseDriverType(db) {
	case drivers.POSTGRES:
		query = query.
			Order("DATE_TRUNC('year', date_shot) DESC").
			Order("DATE_TRUNC('month', date_shot) DESC").
			Order("DATE_TRUNC('day', date_shot) DESC").
			Order("albums.title ASC").
			Order("media.date_shot DESC")
	case drivers.SQLITE:
		query = query.
			Order("strftime('%j', media.date_shot) DESC"). // convert to day of year 001-366
			Order("albums.title ASC").
			Order("TIME(media.date_shot) DESC")
	default:
		query = query.
			Order("YEAR(media.date_shot) DESC").
			Order("MONTH(media.date_shot) DESC").
			Order("DAY(media.date_shot) DESC").
			Order("albums.title ASC").
			Order("TIME(media.date_shot) DESC")
	}

	if fromDate != nil {
		query = query.Where("media.date_shot < ?", fromDate)
	}

	if onlyFavorites != nil && *onlyFavorites {
		query = query.Where("media.id IN (?)", db.Table("user_media_data").Select("user_media_data.media_id").Where("user_media_data.user_id = ?", user.ID).Where("user_media_data.favorite"))
	}

	query = models.FormatSQL(query, nil, paginate)

	var media []*models.Media
	if err := query.Find(&media).Error; err != nil {
		return nil, err
	}

	return media, nil
}
