package dataloader

import (
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"gorm.io/gorm"
)

func NewUserFavoriteLoader(db *gorm.DB) *UserFavoritesLoader {
	return &UserFavoritesLoader{
		maxBatch: 100,
		wait:     5 * time.Millisecond,
		fetch: func(keys []*models.UserMediaData) ([]bool, []error) {

			userIDMap := make(map[int]struct{}, len(keys))
			mediaIDMap := make(map[int]struct{}, len(keys))
			for _, key := range keys {
				userIDMap[key.UserID] = struct{}{}
				mediaIDMap[key.MediaID] = struct{}{}
			}

			uniqueUserIDs := make([]int, len(userIDMap))
			uniqueMediaIDs := make([]int, len(mediaIDMap))

			count := 0
			for id := range userIDMap {
				uniqueUserIDs[count] = id
				count++
			}

			count = 0
			for id := range mediaIDMap {
				uniqueMediaIDs[count] = id
				count++
			}

			var userMediaFavorites []*models.UserMediaData
			err := db.Where("user_id IN (?)", uniqueUserIDs).Where("media_id IN (?)", uniqueMediaIDs).Where("favorite = TRUE").Find(&userMediaFavorites).Error
			if err != nil {
				return nil, []error{err}
			}

			result := make([]bool, len(keys))
			result = iterateFavorites(keys, userMediaFavorites, result)

			return result, nil
		},
	}
}

func iterateFavorites(keys []*models.UserMediaData, userMediaFavorites []*models.UserMediaData, result []bool) []bool {
	for i, key := range keys {
		favorite := false
		for _, fav := range userMediaFavorites {
			if fav.UserID == key.UserID && fav.MediaID == key.MediaID {
				favorite = true
				break
			}
		}
		result[i] = favorite
	}
	return result
}
