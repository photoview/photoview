package resolvers

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strings"

	api "github.com/viktorstrate/photoview/api/graphql"
	"github.com/viktorstrate/photoview/api/graphql/auth"
	"github.com/viktorstrate/photoview/api/graphql/models"
	"github.com/viktorstrate/photoview/api/scanner"
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

func (r *photoResolver) Shares(ctx context.Context, obj *models.Photo) ([]*models.ShareToken, error) {
	rows, err := r.Database.Query("SELECT * FROM share_token WHERE photo_id = ?", obj.PhotoID)
	if err != nil {
		return nil, err
	}

	return models.NewShareTokensFromRows(rows)
}

func (r *photoResolver) Downloads(ctx context.Context, obj *models.Photo) ([]*models.PhotoDownload, error) {

	rows, err := r.Database.Query("SELECT * FROM photo_url WHERE photo_id = ?", obj.PhotoID)
	if err != nil {
		return nil, err
	}

	photoUrls, err := models.NewPhotoURLFromRows(rows)
	if err != nil {
		return nil, err
	}

	downloads := make([]*models.PhotoDownload, 0)

	for _, url := range photoUrls {

		var title string
		switch {
		case url.Purpose == models.PhotoOriginal:
			title = "Original"
		case url.Purpose == models.PhotoThumbnail:
			title = "Small"
		case url.Purpose == models.PhotoHighRes:
			title = "Large"
		}

		downloads = append(downloads, &models.PhotoDownload{
			Title:  title,
			Width:  url.Width,
			Height: url.Height,
			URL:    url.URL(),
		})
	}

	return downloads, nil
}

func (r *photoResolver) HighRes(ctx context.Context, obj *models.Photo) (*models.PhotoURL, error) {
	// Try high res first, then
	web_types_questions := strings.Repeat("?,", len(scanner.WebMimetypes))[:len(scanner.WebMimetypes)*2-1]
	args := make([]interface{}, 0)
	args = append(args, obj.PhotoID, models.PhotoHighRes, models.PhotoOriginal)
	for _, webtype := range scanner.WebMimetypes {
		args = append(args, webtype)
	}

	row := r.Database.QueryRow(`
		SELECT * FROM photo_url WHERE photo_id = ? AND
		(
			purpose = ? OR (purpose = ? AND content_type IN (`+web_types_questions+`))
		) LIMIT 1
	`, args...)

	url, err := models.NewPhotoURLFromRow(row)
	if err != nil {
		log.Printf("Error: Could not query highres: %s\n", err)
		return nil, err
	}

	return url, nil
}

func (r *photoResolver) Thumbnail(ctx context.Context, obj *models.Photo) (*models.PhotoURL, error) {
	row := r.Database.QueryRow("SELECT * FROM photo_url WHERE photo_id = ? AND purpose = ?", obj.PhotoID, models.PhotoThumbnail)

	url, err := models.NewPhotoURLFromRow(row)
	if err != nil {
		log.Printf("Error: Could not query thumbnail: %s\n", err)
		return nil, err
	}

	return url, nil
}

func (r *photoResolver) Album(ctx context.Context, obj *models.Photo) (*models.Album, error) {
	panic("not implemented")
}

func (r *photoResolver) Exif(ctx context.Context, obj *models.Photo) (*models.PhotoEXIF, error) {
	row := r.Database.QueryRow("SELECT photo_exif.* FROM photo NATURAL JOIN photo_exif WHERE photo.photo_id = ?", obj.PhotoID)

	exif, err := models.NewPhotoExifFromRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			return nil, err
		}
	}

	return exif, nil
}
