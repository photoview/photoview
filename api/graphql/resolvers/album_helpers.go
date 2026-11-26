package resolvers

// Helper functions for album.go, kept in a separate file (not the
// gqlgen-managed resolver file) since gqlgen's codegen doesn't recognize
// non-resolver top-level declarations and will otherwise comment them out
// on the next `go generate` as "unknown code".

import (
	"context"

	"github.com/photoview/photoview/api/graphql/models"
)

// maxAlbumTreeChildrenIDs bounds one albumTreeChildren request. The ids come
// from the client, so a broad filter against a large library could otherwise
// build an IN clause past the database driver's parameter limit. The tree
// keeps its own, lower bound on how many nodes it expands at once; this is the
// server's backstop against a client that does not.
const maxAlbumTreeChildrenIDs = 500

// authorizedAlbumIDs narrows albumIds to the ones user may see, in the order
// they were asked for.
//
// Access is inherited: a user linked only to a root album owns everything
// below it. That is the rule OwnsAlbum applies to one album - walk from the
// album up to its root, and it is owned if any album on the way, itself
// included, has a user_albums row for the user. Applied per id, it cost up to
// two queries for every album without a direct row, and this query takes up to
// maxAlbumTreeChildrenIDs ids straight from the client. So the whole batch is
// walked at once instead: each lineage row carries the id it started from, and
// the owned starting ids come back from a single round trip.
//
// depth bounds the walk the way it does in AlbumPath. Parents come from the
// directory tree, so a cycle should not exist, but a bad row must not turn an
// authorization check into a query that never ends.
func (r *queryResolver) authorizedAlbumIDs(ctx context.Context, user *models.User, albumIds []int) ([]int, error) {
	if user.Admin {
		return albumIds, nil
	}

	var ownedIDs []int
	if err := r.DB(ctx).Raw(`
		WITH recursive lineage AS (
			SELECT id AS start_id, id AS album_id, parent_album_id, 0 AS depth
				FROM albums WHERE id IN (?)
			UNION ALL
			SELECT lineage.start_id, parent.id, parent.parent_album_id, lineage.depth + 1
				FROM lineage JOIN albums parent ON parent.id = lineage.parent_album_id
				WHERE lineage.depth < 64
		)
		SELECT DISTINCT lineage.start_id FROM lineage
			JOIN user_albums ON user_albums.album_id = lineage.album_id
			WHERE user_albums.user_id = ?
	`, albumIds, user.ID).Scan(&ownedIDs).Error; err != nil {
		return nil, err
	}

	owned := make(map[int]struct{}, len(ownedIDs))
	for _, id := range ownedIDs {
		owned[id] = struct{}{}
	}

	authorized := make([]int, 0, len(ownedIDs))

	for _, albumID := range albumIds {
		if _, found := owned[albumID]; found {
			authorized = append(authorized, albumID)
		}
	}

	return authorized, nil
}
