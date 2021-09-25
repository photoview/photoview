package actions

import (
	"github.com/photoview/photoview/api/graphql/models"
	"gorm.io/gorm"
)

func MyMedia(db *gorm.DB, user *models.User, order *models.Ordering, paginate *models.Pagination) ([]*models.Media, error) {
	if err := user.FillAlbums(db); err != nil {
		return nil, err
	}

	query := db.Where("media.album_id IN (SELECT user_albums.album_id FROM user_albums WHERE user_albums.user_id = ?)", user.ID)
	query = models.FormatSQL(query, order, paginate)

	var media []*models.Media
	if err := query.Find(&media).Error; err != nil {
		return nil, err
	}

	return media, nil
}
