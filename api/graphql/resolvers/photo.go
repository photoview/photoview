package resolvers

import (
	"context"

	api "github.com/viktorstrate/photoview/api/graphql"
	"github.com/viktorstrate/photoview/api/graphql/auth"
	"github.com/viktorstrate/photoview/api/graphql/models"
)

func (r *queryResolver) MyPhotos(ctx context.Context) ([]*models.Photo, error) {
	panic("Not implemented")
}

func (r *queryResolver) Photo(ctx context.Context, id string) (*models.Photo, error) {
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
	panic("not implemented")
	// row := r.Database.QueryRow("SELECT photo_url.* FROM photo, photo_url WHERE photo.photo_id = ? AND photo.original_url = photo_url.url_id", obj.PhotoID)

	// var photoUrl models.PhotoURL
	// if err := row.Scan(&photoUrl.UrlID, &photoUrl.Token, &photoUrl.Width, &photoUrl.Height); err != nil {
	// 	return nil, err
	// }

	// return &photoUrl, nil
}

func (r *photoResolver) Thumbnail(ctx context.Context, obj *models.Photo) (*models.PhotoURL, error) {
	panic("not implemented")
	// row := r.Database.QueryRow("SELECT photo_url.* FROM photo, photo_url WHERE photo.photo_id = ? AND photo.thumbnail_url = photo_url.url_id", obj.PhotoID)

	// var photoUrl models.PhotoURL
	// if err := row.Scan(&photoUrl.UrlID, &photoUrl.Token, &photoUrl.Width, &photoUrl.Height); err != nil {
	// 	return nil, err
	// }

	// return &photoUrl, nil
}

func (r *photoResolver) Album(ctx context.Context, obj *models.Photo) (*models.Album, error) {
	panic("not implemented")
}

func (r *photoResolver) Exif(ctx context.Context, obj *models.Photo) (*models.PhotoExif, error) {
	panic("not implemented")
}
