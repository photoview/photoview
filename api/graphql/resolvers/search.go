package resolvers

import (
	"context"
	"log"

	"github.com/viktorstrate/photoview/api/graphql/auth"
	"github.com/viktorstrate/photoview/api/graphql/models"
)

func (r *Resolver) Search(ctx context.Context, query string, _limitPhotos *int, _limitAlbums *int) (*models.SearchResult, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	limitPhotos := 10
	limitAlbums := 10

	if _limitPhotos != nil {
		limitPhotos = *_limitPhotos
	}

	if _limitAlbums != nil {
		limitAlbums = *_limitAlbums
	}

	wildQuery := "%" + query + "%"

	photoRows, err := r.Database.Query(`
		SELECT photo.* FROM photo JOIN album ON photo.album_id = album.album_id
		WHERE album.owner_id = ? AND photo.title LIKE ? OR photo.path LIKE ?
		ORDER BY (
			case when photo.title LIKE ? then 2
			     when photo.path LIKE ? then 1
			end ) DESC
		LIMIT ?
	`, user.UserID, wildQuery, wildQuery, wildQuery, wildQuery, limitPhotos)
	if err != nil {
		log.Printf("ERROR: searching photos %s", err)
		return nil, err
	}

	photos, err := models.NewPhotosFromRows(photoRows)
	if err != nil {
		return nil, err
	}

	albumRows, err := r.Database.Query(`
		SELECT * FROM album
		WHERE owner_id = ? AND title LIKE ? OR path LIKE ?
		ORDER BY (
			case when title LIKE ? then 2
			     when path LIKE ? then 1
			end ) DESC
		LIMIT ?
	`, user.UserID, wildQuery, wildQuery, wildQuery, wildQuery, limitAlbums)
	if err != nil {
		log.Printf("ERROR: searching albums %s", err)
		return nil, err
	}

	albums, err := models.NewAlbumsFromRows(albumRows)
	if err != nil {
		return nil, err
	}

	result := models.SearchResult{
		Query:  query,
		Photos: photos,
		Albums: albums,
	}

	return &result, nil
}
