package resolvers

import (
	"context"
	"errors"

	api "github.com/viktorstrate/photoview/api/graphql"
	"github.com/viktorstrate/photoview/api/graphql/auth"
	"github.com/viktorstrate/photoview/api/graphql/models"
)

func (r *queryResolver) MyPhotos(ctx context.Context, filter *models.Filter) ([]*models.Photo, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, errors.New("unauthorized")
	}

	filterSQL, err := filter.FormatSQL()
	if err != nil {
		return nil, err
	}

	rows, err := r.Database.Query("SELECT photo.* FROM photo, album WHERE photo.album_id = album.album_id AND album.owner_id = ?"+filterSQL, user.UserID)
	if err != nil {
		return nil, err
	}

	return models.NewPhotosFromRows(rows)
}

func (r *queryResolver) Photo(ctx context.Context, id int) (*models.Photo, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	row := r.Database.QueryRow(`
		SELECT photo.* FROM photo
		LEFT JOIN album ON photo.album_id = album.album_id
		WHERE photo.photo_id = ? AND album.owner_id = ?
	`, id, user.UserID)

	photo, err := models.NewPhotoFromRow(row)
	if err != nil {
		return nil, err
	}

	return photo, nil
}

type photoResolver struct {
	*Resolver
}

func (r *Resolver) Photo() api.PhotoResolver {
	return &photoResolver{r}
}

func (r *photoResolver) HighRes(ctx context.Context, obj *models.Photo) (*models.PhotoURL, error) {
	panic("not implemented")
}

func (r *photoResolver) Original(ctx context.Context, obj *models.Photo) (*models.PhotoURL, error) {
	row := r.Database.QueryRow("SELECT * FROM photo_url WHERE photo_id = ? AND purpose = ?", obj.PhotoID, models.PhotoOriginal)

	url, err := models.NewPhotoURLFromRow(row)
	if err != nil {
		return nil, err
	}

	return url, nil
}

func (r *photoResolver) Thumbnail(ctx context.Context, obj *models.Photo) (*models.PhotoURL, error) {
	row := r.Database.QueryRow("SELECT * FROM photo_url WHERE photo_id = ? AND purpose = ?", obj.PhotoID, models.PhotoThumbnail)

	url, err := models.NewPhotoURLFromRow(row)
	if err != nil {
		return nil, err
	}

	return url, nil
}

func (r *photoResolver) Album(ctx context.Context, obj *models.Photo) (*models.Album, error) {
	panic("not implemented")
}

func (r *photoResolver) Exif(ctx context.Context, obj *models.Photo) (*models.PhotoExif, error) {
	panic("not implemented")
}
