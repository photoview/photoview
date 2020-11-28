package resolvers

import (
	"context"
	"fmt"
	"log"

	api "github.com/viktorstrate/photoview/api/graphql"
	"github.com/viktorstrate/photoview/api/graphql/auth"
	"github.com/viktorstrate/photoview/api/graphql/models"
)

func (r *queryResolver) MyAlbums(ctx context.Context, filter *models.Filter, onlyRoot *bool, showEmpty *bool, onlyWithFavorites *bool) ([]*models.Album, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	query := r.Database.Where("owner_id = ?", user.ID)

	if onlyRoot != nil && *onlyRoot == true {
		query = query.Where("parent_album = ()", query.Model(&models.Album{})).Select("id").Where("parent_album IS NULL AND owner_id = ?", user.ID)
	}

	if showEmpty == nil || *showEmpty == false {
		subQuery := r.Database.Model(&models.Media{}).Where("album_id = albums.id")

		if onlyWithFavorites != nil && *onlyWithFavorites == true {
			subQuery = subQuery.Where("favorite = 1")
		}

		query = query.Where("EXISTS (?)", subQuery)
	}

	// TODO: Incorporate models.FormatSQL

	var albums []*models.Album
	if err := query.Find(&albums).Error; err != nil {
		return nil, err
	}

	return albums, nil
}

func (r *queryResolver) Album(ctx context.Context, id int) (*models.Album, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	var album models.Album
	if err := r.Database.Where("owner_id = ?", user.ID).First(&album, id).Error; err != nil {
		return nil, err
	}

	return &album, nil
}

func (r *Resolver) Album() api.AlbumResolver {
	return &albumResolver{r}
}

type albumResolver struct{ *Resolver }

func (r *albumResolver) Media(ctx context.Context, album *models.Album, filter *models.Filter, onlyFavorites *bool) ([]*models.Media, error) {

	query := r.Database.
		Joins("Album").
		Where("Album.id = ?", album.ID).
		Where("media.id IN (?)", r.Database.Model(&models.MediaURL{})).Select("media_id").Where("media_url.media_id = media.id")

	if onlyFavorites != nil && *onlyFavorites == true {
		query = query.Where("media.favorite = 1")
	}

	// TODO: Incorporate filter.FormatSQL

	var media []*models.Media
	if err := query.Find(&media).Error; err != nil {
		return nil, err
	}

	return media, nil
}

func (r *albumResolver) Thumbnail(ctx context.Context, obj *models.Album) (*models.Media, error) {

	log.Println("TODO: Album thumbnail migrated yet")

	return nil, nil

	// row := r.Database.QueryRow(`
	// 	WITH recursive sub_albums AS (
	// 		SELECT * FROM album AS root WHERE album_id = ?
	// 		UNION ALL
	// 		SELECT child.* FROM album AS child JOIN sub_albums ON child.parent_album = sub_albums.album_id
	// 	)

	// 	SELECT * FROM media WHERE media.album_id IN (
	// 		SELECT album_id FROM sub_albums
	// 	) AND media.media_id IN (
	// 		SELECT media_id FROM media_url WHERE media_url.media_id = media.media_id
	// 	) LIMIT 1
	// `, obj.AlbumID)

	// media, err := models.NewMediaFromRow(row)
	// if err != nil {
	// 	if err == sql.ErrNoRows {
	// 		return nil, nil
	// 	} else {
	// 		return nil, err
	// 	}
	// }

	// return media, nil
}

func (r *albumResolver) SubAlbums(ctx context.Context, parent *models.Album, filter *models.Filter) ([]*models.Album, error) {

	var albums []*models.Album
	if err := r.Database.Where("parent_album = ?", parent.ID).Find(&albums).Error; err != nil {
		return nil, err
	}

	// TODO: Incorporate filter.FormatSQL

	return albums, nil
}

func (r *albumResolver) ParentAlbum(ctx context.Context, obj *models.Album) (*models.Album, error) {
	panic("not implemented")
}

func (r *albumResolver) Owner(ctx context.Context, obj *models.Album) (*models.User, error) {
	panic("not implemented")
}

func (r *albumResolver) Shares(ctx context.Context, album *models.Album) ([]*models.ShareToken, error) {

	var shareTokens []*models.ShareToken
	if err := r.Database.Where("album_id = ?", album.ID).Find(&shareTokens).Error; err != nil {
		return nil, err
	}

	return shareTokens, nil
}

func (r *albumResolver) Path(ctx context.Context, obj *models.Album) ([]*models.Album, error) {

	fmt.Println("TODO: Album path not migrated yet")

	return make([]*models.Album, 0), nil
	// user := auth.UserFromContext(ctx)
	// if user == nil {
	// 	empty := make([]*models.Album, 0)
	// 	return empty, nil
	// }

	// rows, err := r.Database.Query(`
	// 	WITH recursive path_albums AS (
	// 		SELECT * FROM album anchor WHERE anchor.album_id = ?
	// 		UNION
	// 		SELECT parent.* FROM path_albums child JOIN album parent ON parent.album_id = child.parent_album
	// 	)
	// 	SELECT * FROM path_albums WHERE album_id != ? AND owner_id = ?
	// `, obj.AlbumID, obj.AlbumID, user.UserID)
	// if err != nil {
	// 	return nil, err
	// }

	// return models.NewAlbumsFromRows(rows)
}
