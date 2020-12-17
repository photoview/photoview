package resolvers

import (
	"context"
	"database/sql"

	api "github.com/photoview/photoview/api/graphql"
	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
)

func (r *queryResolver) MyAlbums(ctx context.Context, filter *models.Filter, onlyRoot *bool, showEmpty *bool, onlyWithFavorites *bool) ([]*models.Album, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	filterSQL, err := filter.FormatSQL("album")
	if err != nil {
		return nil, err
	}

	var rows *sql.Rows

	filterFavorites := " AND favorite = 1"
	if onlyWithFavorites == nil || *onlyWithFavorites == false {
		filterFavorites = ""
	}

	filterEmpty := " AND EXISTS (SELECT * FROM media WHERE album_id = album.album_id" + filterFavorites + ") "
	if showEmpty != nil && *showEmpty == true && (onlyWithFavorites == nil || *onlyWithFavorites == false) {
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

func (r *albumResolver) Media(ctx context.Context, obj *models.Album, filter *models.Filter, onlyFavorites *bool) ([]*models.Media, error) {

	filterSQL, err := filter.FormatSQL("media")
	if err != nil {
		return nil, err
	}

	filterFavorites := " AND media.favorite = 1 "
	if onlyFavorites == nil || *onlyFavorites == false {
		filterFavorites = ""
	}

	mediaRows, err := r.Database.Query(`
		SELECT media.* FROM album, media
		WHERE album.album_id = ? AND media.album_id = album.album_id
		AND media.media_id IN (
			SELECT media_id FROM media_url WHERE media_url.media_id = media.media_id
		)
	`+filterFavorites+filterSQL, obj.AlbumID)
	if err != nil {
		return nil, err
	}
	defer mediaRows.Close()

	media, err := models.NewMediaFromRows(mediaRows)
	if err != nil {
		return nil, err
	}

	return media, nil
}

func (r *albumResolver) Thumbnail(ctx context.Context, obj *models.Album) (*models.Media, error) {

	row := r.Database.QueryRow(`
		WITH recursive sub_albums AS (
			SELECT * FROM album AS root WHERE album_id = ?
			UNION ALL
			SELECT child.* FROM album AS child JOIN sub_albums ON child.parent_album = sub_albums.album_id
		)

		SELECT * FROM media WHERE media.album_id IN (
			SELECT album_id FROM sub_albums
		) AND media.media_id IN (
			SELECT media_id FROM media_url WHERE media_url.media_id = media.media_id
		) LIMIT 1
	`, obj.AlbumID)

	media, err := models.NewMediaFromRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			return nil, err
		}
	}

	return media, nil
}

func (r *albumResolver) SubAlbums(ctx context.Context, obj *models.Album, filter *models.Filter) ([]*models.Album, error) {
	filterSQL, err := filter.FormatSQL("album")
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

func (r *albumResolver) Path(ctx context.Context, obj *models.Album) ([]*models.Album, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		empty := make([]*models.Album, 0)
		return empty, nil
	}

	rows, err := r.Database.Query(`
		WITH recursive path_albums AS (
			SELECT * FROM album anchor WHERE anchor.album_id = ?
			UNION
			SELECT parent.* FROM path_albums child JOIN album parent ON parent.album_id = child.parent_album
		)
		SELECT * FROM path_albums WHERE album_id != ? AND owner_id = ?
	`, obj.AlbumID, obj.AlbumID, user.UserID)
	if err != nil {
		return nil, err
	}

	return models.NewAlbumsFromRows(rows)
}
