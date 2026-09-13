package resolvers

// Helper functions for album.go, kept in a separate file (not the
// gqlgen-managed resolver file) since gqlgen's codegen doesn't recognize
// non-resolver top-level declarations and will otherwise comment them out
// on the next `go generate` as "unknown code".

import (
	"context"
	"errors"

	"github.com/photoview/photoview/api/graphql/models"
	"gorm.io/gorm"
)

// maxAlbumTreeChildrenIDs bounds one albumTreeChildren request. The ids come
// from the client, so a broad filter against a large library could otherwise
// build an IN clause past the database driver's parameter limit. The tree
// keeps its own, lower bound on how many nodes it expands at once; this is the
// server's backstop against a client that does not.
const maxAlbumTreeChildrenIDs = 500

// authorizedAlbumIDs narrows albumIds to the ones user may see. Access is
// inherited, so a direct user_albums row is the common case but not the only
// one: a user linked only to a root album owns everything below it. The set
// lookup answers the common case without a query per album, and only the ids
// it misses fall back to the ancestor walk OwnsAlbum performs.
func (r *queryResolver) authorizedAlbumIDs(ctx context.Context, user *models.User, albumIds []int) ([]int, error) {
	if user.Admin {
		return albumIds, nil
	}

	db := r.DB(ctx)

	var directIDs []int
	if err := db.Table("user_albums").
		Where("user_id = ? AND album_id IN (?)", user.ID, albumIds).
		Pluck("album_id", &directIDs).Error; err != nil {
		return nil, err
	}

	direct := make(map[int]struct{}, len(directIDs))
	for _, id := range directIDs {
		direct[id] = struct{}{}
	}

	authorized := make([]int, 0, len(albumIds))

	for _, albumID := range albumIds {
		if _, found := direct[albumID]; found {
			authorized = append(authorized, albumID)

			continue
		}

		var album models.Album
		if err := db.First(&album, albumID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}

			return nil, err
		}

		owns, err := user.OwnsAlbum(db, &album)
		if err != nil {
			return nil, err
		}

		if owns {
			authorized = append(authorized, albumID)
		}
	}

	return authorized, nil
}
