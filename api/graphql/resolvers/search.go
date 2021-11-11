package resolvers

import (
	"context"

	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models/actions"

	"github.com/photoview/photoview/api/graphql/models"
)

func (r *Resolver) Search(ctx context.Context, query string, limitMedia *int, limitAlbums *int) (*models.SearchResult, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	return actions.Search(r.DB(ctx), query, user.ID, limitMedia, limitAlbums)
}
