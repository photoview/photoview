package dataloader

import (
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/media_type"
)

func makeMediaURLLoader(repo MediaRepository, filter func(query *gorm.DB) *gorm.DB) func(keys []int) ([]*models.MediaURL, []error) {
	return func(mediaIDs []int) ([]*models.MediaURL, []error) {
		urls, err := repo.FindMediaURLs(filter, mediaIDs)
		if err != nil {
			return nil, []error{err}
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

func NewThumbnailMediaURLLoader(repo MediaRepository) *MediaURLLoader {
	return &MediaURLLoader{
		maxBatch: 100,
		wait:     5 * time.Millisecond,
		fetch: makeMediaURLLoader(repo, func(query *gorm.DB) *gorm.DB {
			return query.Where("purpose = ? OR purpose = ?", models.PhotoThumbnail, models.VideoThumbnail)
		}),
	}
}

func NewHighresMediaURLLoader(repo MediaRepository) *MediaURLLoader {
	return &MediaURLLoader{
		maxBatch: 100,
		wait:     5 * time.Millisecond,
		fetch: makeMediaURLLoader(repo, func(query *gorm.DB) *gorm.DB {
			return query.Where("purpose = ? OR (purpose = ? AND content_type IN ?)", models.PhotoHighRes, models.MediaOriginal, media_type.WebMimetypes)
		}),
	}
}

func NewVideoWebMediaURLLoader(repo MediaRepository) *MediaURLLoader {
	return &MediaURLLoader{
		maxBatch: 100,
		wait:     5 * time.Millisecond,
		fetch: makeMediaURLLoader(repo, func(query *gorm.DB) *gorm.DB {
			return query.Where("purpose = ? OR purpose = ?", models.VideoWeb, models.MediaOriginal)
		}),
	}
}
