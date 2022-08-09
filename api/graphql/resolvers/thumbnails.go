package resolvers

import (
	"context"

	"github.com/photoview/photoview/api/graphql/models"
	// "github.com/pkg/errors"
	"gorm.io/gorm"
)

func (r *mutationResolver) SetThumbnailDownsampleMethod(ctx context.Context, method models.ThumbnailFilter) (models.ThumbnailFilter, error) {
	db := r.DB(ctx)

	// if method > 5 {
	// 	return 0, errors.New("The requested filter is unsupported, defaulting to nearest neighbor")
	// }

	if err := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Model(&models.SiteInfo{}).Update("thumbnail_method", method).Error; err != nil {
		return models.ThumbnailFilterNearestNeighbor, err
	}

	var siteInfo models.SiteInfo
	if err := db.First(&siteInfo).Error; err != nil {
		return models.ThumbnailFilterNearestNeighbor, err
	}

	return siteInfo.ThumbnailMethod, nil

	// var langTrans *models.LanguageTranslation = nil
	// if language != nil {
	// 	lng := models.LanguageTranslation(*language)
	// 	langTrans = &lng
	// }


}
