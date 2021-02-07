package dataloader

import (
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"gorm.io/gorm"
)

func NewUserLoaderByToken(db *gorm.DB) *UserLoader {
	return &UserLoader{
		maxBatch: 100,
		wait:     5 * time.Millisecond,
		fetch: func(tokens []string) ([]*models.User, []error) {

			var accessTokens []*models.AccessToken
			err := db.Where("expire > ?", time.Now()).Where("value IN (?)", tokens).Find(&accessTokens).Error
			if err != nil {
				return nil, []error{err}
			}

			rows, err := db.Table("access_tokens").Select("distinct user_id").Where("expire > ?", time.Now()).Where("value IN (?)", tokens).Rows()
			if err != nil {
				return nil, []error{err}
			}
			userIDs := make([]int, 0)
			for rows.Next() {
				var id int
				if err := db.ScanRows(rows, &id); err != nil {
					return nil, []error{err}
				}
				userIDs = append(userIDs, id)
			}
			rows.Close()

			var users []*models.User
			if err := db.Where("id IN (?)", userIDs).Find(&users).Error; err != nil {
				return nil, []error{err}
			}

			userMap := make(map[int]*models.User, len(users))
			for _, user := range users {
				userMap[user.ID] = user
			}

			tokenMap := make(map[string]*models.AccessToken, len(tokens))
			for _, token := range accessTokens {
				tokenMap[token.Value] = token
			}

			result := make([]*models.User, len(tokens))
			for i, token := range tokens {
				accessToken, tokenFound := tokenMap[token]
				user, userFound := userMap[accessToken.UserID]
				if tokenFound && userFound {
					result[i] = user
				} else {
					result[i] = nil
				}
			}

			return result, nil
		},
	}
}
