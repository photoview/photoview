package resolvers

import (
	"context"

	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/pkg/errors"
	"gorm.io/gorm/clause"

	"github.com/photoview/photoview/api/graphql/models"
)

func (r *Resolver) Search(ctx context.Context, query string, _limitMedia *int, _limitAlbums *int) (*models.SearchResult, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	limitMedia := 10
	limitAlbums := 10

	if _limitMedia != nil {
		limitMedia = *_limitMedia
	}

	if _limitAlbums != nil {
		limitAlbums = *_limitAlbums
	}

	wildQuery := "%" + query + "%"

	var media []*models.Media

	err := r.Database.Joins("Album").
		Where("Album.owner_id = ? AND ( media.title LIKE ? OR media.path LIKE ? )", user.ID, wildQuery, wildQuery).
		Clauses(clause.OrderBy{
			Expression: clause.Expr{
				SQL:                "(CASE WHEN media.title LIKE ? THEN 2 WHEN media.path LIKE ? THEN 1 END) DESC",
				Vars:               []interface{}{wildQuery, wildQuery},
				WithoutParentheses: true},
		}).
		Limit(limitMedia).
		Find(&media).Error

	if err != nil {
		return nil, errors.Wrapf(err, "searching media")
	}

	var albums []*models.Album

	err = r.Database.Where("owner_id = ? AND (title LIKE ? OR path LIKE ?)", user.ID, wildQuery, wildQuery).
		Clauses(clause.OrderBy{
			Expression: clause.Expr{
				SQL:                "(CASE WHEN title LIKE ? THEN 2 WHEN path LIKE ? THEN 1 END) DESC",
				Vars:               []interface{}{wildQuery, wildQuery},
				WithoutParentheses: true},
		}).
		Limit(limitAlbums).
		Find(&albums).Error

	if err != nil {
		return nil, errors.Wrapf(err, "searching albums")
	}

	result := models.SearchResult{
		Query:  query,
		Media:  media,
		Albums: albums,
	}

	return &result, nil
}
