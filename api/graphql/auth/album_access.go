package auth

import (
	"context"
	"sync"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// albumAccessContextKey stores album membership for one GraphQL operation.
// Media.album, Media.faces and Media.shares otherwise ask once per media item.
// The cache must not be attached to a websocket connection context.
type albumAccessContextKey struct{}

type albumAccessCache struct {
	mu     sync.Mutex
	byUser map[int]map[int]struct{}
}

// WithAlbumAccessCache returns a child context with an empty membership cache.
// GraphQL calls it once per operation. A new operation on the same connection
// context gets a new cache, so a membership change is visible on the next operation.
func WithAlbumAccessCache(ctx context.Context) context.Context {
	return context.WithValue(ctx, albumAccessContextKey{}, &albumAccessCache{})
}

// UserOwnsMedia reports whether media is in an album the user owns, the same
// rule as User.OwnsMedia. Inside one WithAlbumAccessCache context the album IDs
// are loaded once. A context without that cache checks the database every time.
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
