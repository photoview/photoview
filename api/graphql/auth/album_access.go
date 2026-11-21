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
	byUser map[int]albumMembership
}

// albumMembership is the album set for one user during one operation.
// ids is sorted once and then reused. Callers must not modify it.
type albumMembership struct {
	set map[int]struct{}
	ids []int
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
	membership, err := cache.membership(db, user)
	if err != nil {
		return false, err
	}

	_, owned := membership.set[media.AlbumID]
	return owned, nil
}

// UserAlbumIDs returns the albums the user currently belongs to.
// Inside one WithAlbumAccessCache context the membership is loaded once and the
// same sorted slice is returned on every call. Callers must not modify it.
// A context without that cache reads user_albums directly. The IDs are never
// taken from User.Albums, which a websocket connection keeps across operations.
func UserAlbumIDs(ctx context.Context, db *gorm.DB, user *models.User) ([]int, error) {
	if cache, ok := ctx.Value(albumAccessContextKey{}).(*albumAccessCache); ok && cache != nil {
		membership, err := cache.membership(db, user)
		if err != nil {
			return nil, err
		}
		return membership.ids, nil
	}

	set, err := loadUserAlbumIDs(db, user.ID)
	if err != nil {
		return nil, err
	}
	return sortedAlbumIDs(set), nil
}

func (cache *albumAccessCache) membership(db *gorm.DB, user *models.User) (albumMembership, error) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if cache.byUser == nil {
		cache.byUser = make(map[int]albumMembership)
	}

	membership, ok := cache.byUser[user.ID]
	if ok {
		return membership, nil
	}

	loaded, err := loadUserAlbumIDs(db, user.ID)
	if err != nil {
		return albumMembership{}, err
	}
	membership = albumMembership{set: loaded, ids: sortedAlbumIDs(loaded)}
	cache.byUser[user.ID] = membership
	return membership, nil
}

func sortedAlbumIDs(set map[int]struct{}) []int {
	ids := make([]int, 0, len(set))
	for id := range set {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	return ids
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
