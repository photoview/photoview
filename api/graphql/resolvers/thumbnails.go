package resolvers

import (
	"context"
	
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func (r *mutationResolver) SetThumbnailDownsampleMethod(ctx context.Context, method int) (int, error) {
	db := r.DB(ctx)

	if method > 5 {
		return 0, errors.New("The requested filter is unsupported, defaulting to nearest neighbor")
	}

	if err := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Model(&models.SiteInfo{}).Update("thumbnail_method", method).Error; err != nil {
		return 0, err
	}

	var siteInfo models.SiteInfo
	if err := db.First(&siteInfo).Error; err != nil {
		return 0, err
	}

	// scanner_queue.ChangeScannerConcurrentWorkers(siteInfo.ConcurrentWorkers)

	return siteInfo.ThumbnailMethod, nil
}
