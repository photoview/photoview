package resolvers

import (
	"context"
	"fmt"
	"time"

	"github.com/photoview/photoview/api/database"
	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
	"gorm.io/gorm"
)

func (r *queryResolver) MyTimeline(ctx context.Context, paginate *models.Pagination, onlyFavorites *bool) ([]*models.TimelineGroup, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	var timelineGroups []*models.TimelineGroup

	transactionError := r.Database.Transaction(func(tx *gorm.DB) error {
		// album_id, year, month, day
		daysQuery := tx.Select(
			"albums.id AS album_id",
			fmt.Sprintf("%s AS year", database.DateExtract(tx, database.DateCompYear, "media.date_shot")),
			fmt.Sprintf("%s AS month", database.DateExtract(tx, database.DateCompMonth, "media.date_shot")),
			fmt.Sprintf("%s AS day", database.DateExtract(tx, database.DateCompDay, "media.date_shot")),
		).
			Table("media").
			Joins("JOIN albums ON media.album_id = albums.id").
			Where("albums.id IN (?)", tx.Table("user_albums").Select("user_albums.album_id").Where("user_id = ?", user.ID))

		if onlyFavorites != nil && *onlyFavorites == true {
			daysQuery.Where("media.id IN (?)", tx.Table("user_media_data").Select("user_media_data.media_id").Where("user_media_data.user_id = ?", user.ID).Where("user_media_data.favorite = 1"))
		}

		if paginate != nil {
			if paginate.Limit != nil {
				daysQuery.Limit(*paginate.Limit)
			}

			if paginate.Offset != nil {
				daysQuery.Offset(*paginate.Offset)
			}
		}

		dateChain := fmt.Sprintf("%s, %s, %s",
			database.DateExtract(tx, database.DateCompYear, "media.date_shot"),
			database.DateExtract(tx, database.DateCompMonth, "media.date_shot"),
			database.DateExtract(tx, database.DateCompDay, "media.date_shot"))

		rows, err := daysQuery.Group("albums.id").Group(dateChain).Order(dateChain).Rows()

		defer rows.Close()

		if err != nil {
			return err
		}

		type group struct {
			albumID int
			year    int
			month   int
			day     int
		}

		dbGroups := make([]group, 0)

		for rows.Next() {
			var g group
			rows.Scan(&g.albumID, &g.year, &g.month, &g.day)
			dbGroups = append(dbGroups, g)
		}

		timelineGroups = make([]*models.TimelineGroup, len(dbGroups))

		for i, group := range dbGroups {

			// Fill album
			var groupAlbum models.Album
			if err := tx.First(&groupAlbum, group.albumID).Error; err != nil {
				return err
			}

			// Fill media
			var groupMedia []*models.Media
			mediaQuery := tx.Model(&models.Media{}).
				Where("album_id = ?", group.albumID).
				Where(fmt.Sprintf("%s = ?", database.DateExtract(tx, database.DateCompYear, "media.date_shot")), group.year).
				Where(fmt.Sprintf("%s = ?", database.DateExtract(tx, database.DateCompMonth, "media.date_shot")), group.month).
				Where(fmt.Sprintf("%s = ?", database.DateExtract(tx, database.DateCompDay, "media.date_shot")), group.day).
				Order("date_shot DESC")

			if onlyFavorites != nil && *onlyFavorites == true {
				mediaQuery.Where("media.id IN (?)", tx.Table("user_media_data").Select("user_media_data.media_id").Where("user_media_data.user_id = ?", user.ID).Where("user_media_data.favorite = 1"))
			}

			if err := mediaQuery.Limit(5).Find(&groupMedia).Error; err != nil {
				return err
			}

			// Get total media count
			var totalMedia int64
			if err := mediaQuery.Count(&totalMedia).Error; err != nil {
				return err
			}

			var date time.Time = groupMedia[0].DateShot
			date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

			timelineGroup := models.TimelineGroup{
				Album:      &groupAlbum,
				Media:      groupMedia,
				MediaTotal: int(totalMedia),
				Date:       date,
			}

			timelineGroups[i] = &timelineGroup
		}

		return nil
	})

	if transactionError != nil {
		return nil, transactionError
	}

	return timelineGroups, nil
}
