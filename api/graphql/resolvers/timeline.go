package resolvers

import (
	"context"
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"gorm.io/gorm"
)

func (r *queryResolver) MyTimeline(ctx context.Context) ([]*models.TimelineGroup, error) {

	var timelineGroups []*models.TimelineGroup

	transactionError := r.Database.Transaction(func(tx *gorm.DB) error {
		// album_id, year, month, day
		rows, err := tx.Select(
			"albums.id AS album_id",
			"YEAR(media.date_shot) AS year",
			"MONTH(media.date_shot) AS month",
			"DAY(media.date_shot) AS day",
		).
			Table("media").
			Joins("JOIN albums ON media.album_id = albums.id").
			Group("albums.id, YEAR(media.date_shot), MONTH(media.date_shot), DAY(media.date_shot)").
			Order("media.date_shot DESC").
			Rows()

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
			err := tx.Model(&models.Media{}).
				Where("album_id = ? AND YEAR(date_shot) = ? AND MONTH(date_shot) = ? AND DAY(date_shot) = ?", group.albumID, group.year, group.month, group.day).
				Order("date_shot DESC").
				Limit(5).
				Find(&groupMedia).Error

			if err != nil {
				return err
			}

			// Get total media count
			var totalMedia int64
			err = tx.Model(&models.Media{}).
				Where("album_id = ? AND YEAR(date_shot) = ? AND MONTH(date_shot) = ? AND DAY(date_shot) = ?", group.albumID, group.year, group.month, group.day).
				Count(&totalMedia).Error

			if err != nil {
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
