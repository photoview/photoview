package dataloader

import (
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"gorm.io/gorm"
)

func NewAlbumHiddenLoader(db *gorm.DB) *AlbumHiddenLoader {
	return &AlbumHiddenLoader{
		maxBatch: 100,
		wait:     5 * time.Millisecond,
		fetch: func(keys []models.UserAlbumKey) ([]bool, []error) {

			userIDMap := make(map[int]struct{}, len(keys))
			albumIDMap := make(map[int]struct{}, len(keys))
			for _, key := range keys {
				userIDMap[key.UserID] = struct{}{}
				albumIDMap[key.AlbumID] = struct{}{}
			}

			uniqueUserIDs := make([]int, len(userIDMap))
			uniqueAlbumIDs := make([]int, len(albumIDMap))

			count := 0
			for id := range userIDMap {
				uniqueUserIDs[count] = id
				count++
			}

			count = 0
			for id := range albumIDMap {
				uniqueAlbumIDs[count] = id
				count++
			}

			var hiddenRows []*models.UserAlbumData
			err := db.Where("user_id IN (?)", uniqueUserIDs).Where("album_id IN (?)", uniqueAlbumIDs).Where("hidden = TRUE").Find(&hiddenRows).Error
			if err != nil {
				return nil, []error{err}
			}

			result := make([]bool, len(keys))
			for i, key := range keys {
				hidden := false
				for _, row := range hiddenRows {
					if row.UserID == key.UserID && row.AlbumID == key.AlbumID {
						hidden = true
						break
					}
				}
				result[i] = hidden
			}

			return result, nil
		},
	}
}
