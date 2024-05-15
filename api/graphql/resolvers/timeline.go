package resolvers

import (
	"context"
	"time"

	"github.com/kkovaletp/photoview/api/graphql/auth"
	"github.com/kkovaletp/photoview/api/graphql/models"
	"github.com/kkovaletp/photoview/api/graphql/models/actions"
)

func (r *queryResolver) MyTimeline(ctx context.Context, paginate *models.Pagination, onlyFavorites *bool, fromDate *time.Time) ([]*models.Media, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	return actions.MyTimeline(r.DB(ctx), user, paginate, onlyFavorites, fromDate)
}
