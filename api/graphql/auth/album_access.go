package auth

import (
	"context"
	"sync"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// albumAccessContextKey stores one album-membership cache for a request.
// Media.album, Media.faces and Media.shares otherwise ask once per media item.
type albumAccessContextKey struct{}

type albumAccessCache struct {
	mu     sync.Mutex
	byUser map[int]map[int]struct{}
}

// UserOwnsMedia reports whether media is in an album the user owns, the same
// rule as User.OwnsMedia. A context from AddUserToContext loads that user's
// album IDs once and reuses them for the rest of the request.
func UserOwnsMedia(ctx context.Context, db *gorm.DB, user *models.User, media *models.Media) (bool, error) {
	if cache, ok := ctx.Value(albumAccessContextKey{}).(*albumAccessCache); ok && cache != nil {
		return cache.owns(db, user, media)
	}
	return user.OwnsMedia(db, media)
}

func (cache *albumAccessCache) owns(db *gorm.DB, user *models.User, media *models.Media) (bool, error) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if cache.byUser == nil {
		cache.byUser = make(map[int]map[int]struct{})
	}

	ids, ok := cache.byUser[user.ID]
	if !ok {
		var memberships []models.UserAlbums
		err := db.Where("user_id = ?", user.ID).Find(&memberships).Error
		if err != nil {
			return false, errors.Wrap(err, "check media album owner in database")
		}

		ids = make(map[int]struct{}, len(memberships))
		for _, membership := range memberships {
			ids[membership.AlbumID] = struct{}{}
		}
		cache.byUser[user.ID] = ids
	}

	_, owned := ids[media.AlbumID]
	return owned, nil
}
