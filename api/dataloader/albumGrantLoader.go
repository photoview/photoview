package dataloader

import (
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"gorm.io/gorm"
)

func NewAlbumGrantLoader(db *gorm.DB) *AlbumGrantLoader {
	return &AlbumGrantLoader{
		maxBatch: 100,
		wait:     5 * time.Millisecond,
		fetch: func(keys []models.UserAlbumKey) ([]*models.UserAlbums, []error) {

			userIDMap := make(map[int]struct{}, len(keys))
			albumIDMap := make(map[int]struct{}, len(keys))
			for _, key := range keys {
				userIDMap[key.UserID] = struct{}{}
				albumIDMap[key.AlbumID] = struct{}{}
			}

			uniqueUserIDs := make([]int, 0, len(userIDMap))
			for id := range userIDMap {
				uniqueUserIDs = append(uniqueUserIDs, id)
			}

			uniqueAlbumIDs := make([]int, 0, len(albumIDMap))
			for id := range albumIDMap {
				uniqueAlbumIDs = append(uniqueAlbumIDs, id)
			}

			var grants []*models.UserAlbums
			err := db.Where("user_id IN (?)", uniqueUserIDs).Where("album_id IN (?)", uniqueAlbumIDs).Find(&grants).Error
			if err != nil {
				return nil, []error{err}
			}

			result := make([]*models.UserAlbums, len(keys))
			for i, key := range keys {
				for _, grant := range grants {
					if grant.UserID == key.UserID && grant.AlbumID == key.AlbumID {
						result[i] = grant
						break
					}
				}
			}

			return result, nil
		},
	}
}
