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

func (r *queryResolver) MyMedia(ctx context.Context, filter *models.Filter) ([]*models.Media, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, errors.New("unauthorized")
	}

	filterSQL, err := filter.FormatSQL()
	if err != nil {
		return nil, err
	}

	rows, err := r.Database.Query(`
		SELECT media.* FROM media, album
		WHERE media.album_id = album.album_id AND album.owner_id = ?
		AND media.media_id IN (
			SELECT media_id FROM media_url WHERE media_url.media_id = media.media_id
		)
	`+filterSQL, user.UserID)
	if err != nil {
		return nil, err
	}

	return models.NewMediaFromRows(rows)
}

func (r *queryResolver) Media(ctx context.Context, id int) (*models.Media, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	row := r.Database.QueryRow(`
		SELECT media.* FROM media
		JOIN album ON media.album_id = album.album_id
		WHERE media.media_id = ? AND album.owner_id = ?
		AND media.media_id IN (
			SELECT media_id FROM media_url WHERE media_url.media_id = media.media_id
		)
	`, id, user.UserID)

	media, err := models.NewMediaFromRow(row)
	if err != nil {
		return nil, err
	}

	return media, nil
}

type mediaResolver struct {
	*Resolver
}

func (r *Resolver) Media() api.MediaResolver {
	return &mediaResolver{r}
}

func (r *mediaResolver) Shares(ctx context.Context, obj *models.Media) ([]*models.ShareToken, error) {
	rows, err := r.Database.Query("SELECT * FROM share_token WHERE media_id = ?", obj.MediaID)
	if err != nil {
		return nil, err
	}

	return models.NewShareTokensFromRows(rows)
}

func (r *mediaResolver) Downloads(ctx context.Context, obj *models.Media) ([]*models.MediaDownload, error) {

	rows, err := r.Database.Query("SELECT * FROM media_url WHERE media_id = ?", obj.MediaID)
	if err != nil {
		return nil, err
	}

	mediaUrls, err := models.NewMediaURLFromRows(rows)
	if err != nil {
		return nil, err
	}

	downloads := make([]*models.MediaDownload, 0)

	for _, url := range mediaUrls {

		var title string
		switch {
		case url.Purpose == models.MediaOriginal:
			title = "Original"
		case url.Purpose == models.PhotoThumbnail:
			title = "Small"
		case url.Purpose == models.PhotoHighRes:
			title = "Large"
		}

		downloads = append(downloads, &models.MediaDownload{
			Title:  title,
			Width:  url.Width,
			Height: url.Height,
			URL:    url.URL(),
		})
	}

	return downloads, nil
}

func (r *mediaResolver) HighRes(ctx context.Context, obj *models.Media) (*models.MediaURL, error) {
	// Try high res first, then
	web_types_questions := strings.Repeat("?,", len(scanner.WebMimetypes))[:len(scanner.WebMimetypes)*2-1]
	args := make([]interface{}, 0)
	args = append(args, obj.MediaID, models.PhotoHighRes, models.MediaOriginal)
	for _, webtype := range scanner.WebMimetypes {
		args = append(args, webtype)
	}

	row := r.Database.QueryRow(`
		SELECT * FROM media_url WHERE media_id = ? AND
		(
			purpose = ? OR (purpose = ? AND content_type IN (`+web_types_questions+`))
		) LIMIT 1
	`, args...)

	url, err := models.NewMediaURLFromRow(row)
	if err != nil {
		log.Printf("Error: Could not query highres: %s\n", err)
		return nil, err
	}

	return url, nil
}

func (r *mediaResolver) Thumbnail(ctx context.Context, obj *models.Media) (*models.MediaURL, error) {
	row := r.Database.QueryRow("SELECT * FROM media_url WHERE media_id = ? AND purpose = ?", obj.MediaID, models.PhotoThumbnail)

	url, err := models.NewMediaURLFromRow(row)
	if err != nil {
		log.Printf("Error: Could not query thumbnail: %s\n", err)
		return nil, err
	}

	return url, nil
}

func (r *mediaResolver) Album(ctx context.Context, obj *models.Media) (*models.Album, error) {
	row := r.Database.QueryRow("SELECT album.* from media JOIN album ON media.album_id = album.album_id WHERE media_id = ?", obj.MediaID)
	return models.NewAlbumFromRow(row)
}

func (r *mediaResolver) Exif(ctx context.Context, obj *models.Media) (*models.MediaEXIF, error) {
	row := r.Database.QueryRow("SELECT media_exif.* FROM media NATURAL JOIN media_exif WHERE media.media_id = ?", obj.MediaID)

	exif, err := models.NewMediaExifFromRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			return nil, err
		}
	}

	return exif, nil
}

func (r *mutationResolver) FavoriteMedia(ctx context.Context, mediaID int, favorite bool) (*models.Media, error) {

	user := auth.UserFromContext(ctx)

	row := r.Database.QueryRow("SELECT media.* FROM media JOIN album ON media.album_id = album.album_id WHERE media.media_id = ? AND album.owner_id = ?", mediaID, user.UserID)

	media, err := models.NewMediaFromRow(row)
	if err != nil {
		return nil, err
	}

	_, err = r.Database.Exec("UPDATE media SET favorite = ? WHERE media_id = ?", favorite, media.MediaID)
	if err != nil {
		return nil, err
	}

	media.Favorite = favorite

	return media, nil
}
