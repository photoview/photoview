package dataloader

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/photoview/photoview/api/graphql/models"
)

type MediaRepository interface {
	FindMediaURLs(query func(*gorm.DB) *gorm.DB, mediaIDs []int) ([]*models.MediaURL, error)
}

type GormMediaRepository struct {
	db *gorm.DB
}

func (r *GormMediaRepository) FindMediaURLs(filter func(*gorm.DB) *gorm.DB, mediaIDs []int) ([]*models.MediaURL, error) {
	var urls []*models.MediaURL
	query := r.db.Where("media_id IN (?)", mediaIDs)
	filter(query)

	if err := query.Find(&urls).Error; err != nil {
		return nil, errors.Wrap(err, "media url loader database query")
	}

	return urls, nil
}