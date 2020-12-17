package resolvers

import (
	"context"

	"github.com/photoview/photoview/api/graphql/models"
)

func (r *Resolver) Search(ctx context.Context, query string, _limitMedia *int, _limitAlbums *int) (*models.SearchResult, error) {
	// user := auth.UserFromContext(ctx)
	// if user == nil {
	// 	return nil, auth.ErrUnauthorized
	// }

	// limitMedia := 10
	// limitAlbums := 10

	// if _limitMedia != nil {
	// 	limitMedia = *_limitMedia
	// }

	// if _limitAlbums != nil {
	// 	limitAlbums = *_limitAlbums
	// }

	// wildQuery := "%" + query + "%"

	// photoRows, err := r.Database.Query(`
	// 	SELECT media.* FROM media JOIN album ON media.album_id = album.album_id
	// 	WHERE album.owner_id = ? AND ( media.title LIKE ? OR media.path LIKE ? )
	// 	ORDER BY (
	// 		case when media.title LIKE ? then 2
	// 		     when media.path LIKE ? then 1
	// 		end ) DESC
	// 	LIMIT ?
	// `, user.UserID, wildQuery, wildQuery, wildQuery, wildQuery, limitMedia)
	// if err != nil {
	// 	return nil, errors.Wrapf(err, "searching media")
	// }

	// photos, err := models.NewMediaFromRows(photoRows)
	// if err != nil {
	// 	return nil, err
	// }

	// albumRows, err := r.Database.Query(`
	// 	SELECT * FROM album
	// 	WHERE owner_id = ? AND ( title LIKE ? OR path LIKE ? )
	// 	ORDER BY (
	// 		case when title LIKE ? then 2
	// 		     when path LIKE ? then 1
	// 		end ) DESC
	// 	LIMIT ?
	// `, user.UserID, wildQuery, wildQuery, wildQuery, wildQuery, limitAlbums)
	// if err != nil {
	// 	return nil, errors.Wrapf(err, "searching albums")
	// }

	// albums, err := models.NewAlbumsFromRows(albumRows)
	// if err != nil {
	// 	return nil, err
	// }

	// result := models.SearchResult{
	// 	Query:  query,
	// 	Media:  photos,
	// 	Albums: albums,
	// }

	// return &result, nil
	panic("to be migrated")
}
