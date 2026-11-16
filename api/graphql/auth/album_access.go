package auth

import (
	"context"
	"sort"
	"sync"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// albumAccessContextKey stores album membership for one GraphQL operation.
// Media.album, Media.faces, Media.shares and the face-group lookups otherwise
// ask once per field. The cache must not be attached to a websocket connection
// context: User.Albums lives on the user value that connection keeps, and
// FillAlbums will not reload it.
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
	ids, err := cache.albumIDs(db, user)
	if err != nil {
		return false, err
	}

	_, owned := ids[media.AlbumID]
	return owned, nil
}

// UserAlbumIDs returns the albums the user currently belongs to.
// Inside one WithAlbumAccessCache context the membership is loaded once.
// A context without that cache reads user_albums directly. The IDs are never
// taken from User.Albums, which a websocket connection keeps across operations.
func UserAlbumIDs(ctx context.Context, db *gorm.DB, user *models.User) ([]int, error) {
	var set map[int]struct{}
	var err error
	if cache, ok := ctx.Value(albumAccessContextKey{}).(*albumAccessCache); ok && cache != nil {
		set, err = cache.albumIDs(db, user)
	} else {
		set, err = loadUserAlbumIDs(db, user.ID)
	}
	if err != nil {
		return nil, err
	}

	ids := make([]int, 0, len(set))
	for id := range set {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	return ids, nil
}

func (cache *albumAccessCache) albumIDs(db *gorm.DB, user *models.User) (map[int]struct{}, error) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if cache.byUser == nil {
		cache.byUser = make(map[int]map[int]struct{})
	}

	ids, ok := cache.byUser[user.ID]
	if ok {
		return ids, nil
	}

	loaded, err := loadUserAlbumIDs(db, user.ID)
	if err != nil {
		return nil, err
	}
	cache.byUser[user.ID] = loaded
	return loaded, nil
}

func loadUserAlbumIDs(db *gorm.DB, userID int) (map[int]struct{}, error) {
	var memberships []models.UserAlbums
	err := db.Where("user_id = ?", userID).Find(&memberships).Error
	if err != nil {
		return nil, errors.Wrap(err, "check media album owner in database")
	}

	ids := make(map[int]struct{}, len(memberships))
	for _, membership := range memberships {
		ids[membership.AlbumID] = struct{}{}
	}
	return ids, nil
}
