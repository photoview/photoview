package dataloader

import (
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/media_type"
	"github.com/pkg/errors"

	"gorm.io/gorm"
)

func makeMediaURLLoader(db *gorm.DB, filter func(query *gorm.DB) *gorm.DB) func(keys []int) ([]*models.MediaURL, []error) {
	return func(mediaIDs []int) ([]*models.MediaURL, []error) {

		var urls []*models.MediaURL
		query := db.Where("media_id IN (?)", mediaIDs)

		query = filter(query)

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
			return query.Where("purpose IN ?", []string{string(models.PhotoThumbnail), string(models.VideoThumbnail)})
		}),
	}
}

func NewHighresMediaURLLoader(db *gorm.DB) *MediaURLLoader {
	return &MediaURLLoader{
		maxBatch: 100,
		wait:     5 * time.Millisecond,
		fetch: makeMediaURLLoader(db, func(query *gorm.DB) *gorm.DB {
			return query.
				Where("(purpose = ? OR (purpose = ? AND content_type IN ?))", models.PhotoHighRes, models.MediaOriginal, media_type.WebMimetypes).
				//PhotoHighRes consistently wins ordering when both exist, which is preferred for web delivery
				Order("media_id ASC, CASE purpose WHEN '" +
					string(models.MediaOriginal) + "' THEN 0 WHEN '" +
					string(models.PhotoHighRes) + "' THEN 1 END ASC")
		}),
	}
}

func NewVideoWebMediaURLLoader(db *gorm.DB) *MediaURLLoader {
	return &MediaURLLoader{
		maxBatch: 100,
		wait:     5 * time.Millisecond,
		fetch: makeMediaURLLoader(db, func(query *gorm.DB) *gorm.DB {
			return query.
				Where("purpose IN ?", []string{string(models.VideoWeb), string(models.MediaOriginal)}).
				//VideoWeb consistently wins ordering when both exist, which is preferred for web delivery
				Order("media_id ASC, CASE purpose WHEN '" +
					string(models.MediaOriginal) + "' THEN 0 WHEN '" +
					string(models.VideoWeb) + "' THEN 1 END ASC")
		}),
	}
}
