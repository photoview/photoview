package resolvers

import (
	"context"
	"database/sql"

	api "github.com/viktorstrate/photoview/api/graphql"
	"github.com/viktorstrate/photoview/api/graphql/auth"
	"github.com/viktorstrate/photoview/api/graphql/models"
)

func (r *queryResolver) MyAlbums(ctx context.Context, filter *models.Filter, onlyRoot *bool, showEmpty *bool) ([]*models.Album, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	filterSQL, err := filter.FormatSQL()
	if err != nil {
		return nil, err
	}

	var rows *sql.Rows

	filterEmpty := " AND EXISTS (SELECT * FROM photo WHERE album_id = album.album_id) "
	if showEmpty != nil && *showEmpty == true {
		filterEmpty = ""
	}

	if onlyRoot == nil || *onlyRoot == false {
		rows, err = r.Database.Query("SELECT * FROM album WHERE owner_id = ?"+filterEmpty+filterSQL, user.UserID)
		if err != nil {
			return nil, err
		}
	} else {
		rows, err = r.Database.Query(`
			SELECT * FROM album WHERE owner_id = ? AND parent_album = (
				SELECT album_id FROM album WHERE parent_album IS NULL AND owner_id = ?
			)
		`+filterEmpty+filterSQL, user.UserID, user.UserID)
		if err != nil {
			return nil, err
		}
	}

	albums, err := models.NewAlbumsFromRows(rows)
	if err != nil {
		return nil, err
	}

	return albums, nil
}

func (r *queryResolver) Album(ctx context.Context, id int) (*models.Album, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	row := r.Database.QueryRow("SELECT * FROM album WHERE album_id = ? AND owner_id = ?", id, user.UserID)
	album, err := models.NewAlbumFromRow(row)
	if err != nil {
		return nil, err
	}

	return album, nil
}

func (r *Resolver) Album() api.AlbumResolver {
	return &albumResolver{r}
}

type albumResolver struct{ *Resolver }

func (r *albumResolver) Photos(ctx context.Context, obj *models.Album, filter *models.Filter) ([]*models.Photo, error) {

	filterSQL, err := filter.FormatSQL()
	if err != nil {
		return nil, err
	}

	photoRows, err := r.Database.Query(`
		SELECT photo.* FROM album, photo
		WHERE album.album_id = ? AND photo.album_id = album.album_id
		AND photo.photo_id IN (
			SELECT photo_id FROM photo_url WHERE photo_url.photo_id = photo.photo_id
		)
	`+filterSQL, obj.AlbumID)
	if err != nil {
		return nil, err
	}
	defer photoRows.Close()

	photos, err := models.NewPhotosFromRows(photoRows)
	if err != nil {
		return nil, err
	}

	return photos, nil
}

func (r *albumResolver) Thumbnail(ctx context.Context, obj *models.Album) (*models.Photo, error) {

	row := r.Database.QueryRow(`
		WITH recursive sub_albums AS (
			SELECT * FROM album AS root WHERE album_id = ?
			UNION ALL
			SELECT child.* FROM album AS child JOIN sub_albums ON child.parent_album = sub_albums.album_id
		)

		SELECT * FROM photo WHERE photo.album_id IN (
			SELECT album_id FROM sub_albums
		) AND photo.photo_id IN (
			SELECT photo_id FROM photo_url WHERE photo_url.photo_id = photo.photo_id
		) LIMIT 1
	`, obj.AlbumID)

	photo, err := models.NewPhotoFromRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			return nil, err
		}
	}

	return photo, nil
}

func (r *albumResolver) SubAlbums(ctx context.Context, obj *models.Album, filter *models.Filter) ([]*models.Album, error) {
	filterSQL, err := filter.FormatSQL()
	if err != nil {
		return nil, err
	}

	rows, err := r.Database.Query("SELECT * FROM album WHERE parent_album = ?"+filterSQL, obj.AlbumID)
	if err != nil {
		return nil, err
	}

	albums, err := models.NewAlbumsFromRows(rows)
	if err != nil {
		return nil, err
	}

	return albums, nil
}

func (r *albumResolver) ParentAlbum(ctx context.Context, obj *models.Album) (*models.Album, error) {
	panic("not implemented")
}

func (r *albumResolver) Owner(ctx context.Context, obj *models.Album) (*models.User, error) {
	panic("not implemented")
}

func (r *albumResolver) Shares(ctx context.Context, obj *models.Album) ([]*models.ShareToken, error) {
	rows, err := r.Database.Query("SELECT * FROM share_token WHERE album_id = ?", obj.ID())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return models.NewShareTokensFromRows(rows)
}
