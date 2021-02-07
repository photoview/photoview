package dataloader

import (
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner"
	"github.com/pkg/errors"

	"gorm.io/gorm"
)

func makeMediaURLLoader(db *gorm.DB, filter func(query *gorm.DB) *gorm.DB) func(keys []int) ([]*models.MediaURL, []error) {
	return func(mediaIDs []int) ([]*models.MediaURL, []error) {

		var urls []*models.MediaURL
		query := db.Where("media_id IN (?)", mediaIDs)

		filter(query)

		if err := query.Find(&urls).Error; err != nil {
			return nil, []error{errors.Wrap(err, "media url loader database query")}
		}

		resultMap := make(map[int]*models.MediaURL, len(mediaIDs))
		for _, url := range urls {
			resultMap[url.MediaID] = url
		}

		result := make([]*models.MediaURL, len(mediaIDs))
		for i, mediaID := range mediaIDs {
			mediaURL, found := resultMap[mediaID]
			if found {
				result[i] = mediaURL
			} else {
				result[i] = nil
			}
		}

		return result, nil
	}
}

func NewThumbnailMediaURLLoader(db *gorm.DB) *MediaURLLoader {
	return &MediaURLLoader{
		maxBatch: 100,
		wait:     5 * time.Millisecond,
		fetch: makeMediaURLLoader(db, func(query *gorm.DB) *gorm.DB {
			return query.Where("purpose = ? OR purpose = ?", models.PhotoThumbnail, models.VideoThumbnail)
		}),
	}
}

func NewHighresMediaURLLoader(db *gorm.DB) *MediaURLLoader {
	return &MediaURLLoader{
		maxBatch: 100,
		wait:     5 * time.Millisecond,
		fetch: makeMediaURLLoader(db, func(query *gorm.DB) *gorm.DB {
			return query.Where("purpose = ? OR (purpose = ? AND content_type IN ?)", models.PhotoHighRes, models.MediaOriginal, scanner.WebMimetypes)
		}),
	}
}

func NewVideoWebMediaURLLoader(db *gorm.DB) *MediaURLLoader {
	return &MediaURLLoader{
		maxBatch: 100,
		wait:     5 * time.Millisecond,
		fetch: makeMediaURLLoader(db, func(query *gorm.DB) *gorm.DB {
			return query.Where("purpose = ? OR purpose = ?", models.VideoWeb, models.MediaOriginal)
		}),
	}
}
