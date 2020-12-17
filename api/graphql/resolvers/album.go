package resolvers

import (
	"context"

	api "github.com/photoview/photoview/api/graphql"
	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
)

func (r *queryResolver) MyAlbums(ctx context.Context, filter *models.Filter, onlyRoot *bool, showEmpty *bool, onlyWithFavorites *bool) ([]*models.Album, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	query := r.Database.Where("owner_id = ?", user.ID)

	if onlyRoot != nil && *onlyRoot == true {
		query = query.Where("parent_album_id = (?)", r.Database.Model(&models.Album{}).Select("id").Where("parent_album_id IS NULL AND owner_id = ?", user.ID))
	}

	if showEmpty == nil || *showEmpty == false {
		subQuery := r.Database.Model(&models.Media{}).Where("album_id = albums.id")

		if onlyWithFavorites != nil && *onlyWithFavorites == true {
			subQuery = subQuery.Where("favorite = 1")
		}

		query = query.Where("EXISTS (?)", subQuery)
	}

	query = filter.FormatSQL(query)

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
		Where("media.id IN (?)", r.Database.Model(&models.MediaURL{}).Select("media_urls.media_id").Where("media_urls.media_id = media.id"))

	if onlyFavorites != nil && *onlyFavorites == true {
		query = query.Where("media.favorite = 1")
	}

	query = filter.FormatSQL(query)

	var media []*models.Media
	if err := query.Find(&media).Error; err != nil {
		return nil, err
	}

	return media, nil
}

func (r *albumResolver) Thumbnail(ctx context.Context, obj *models.Album) (*models.Media, error) {

	var media models.Media

	err := r.Database.Raw(`
		WITH recursive sub_albums AS (
			SELECT * FROM albums AS root WHERE id = ?
			UNION ALL
			SELECT child.* FROM albums AS child JOIN sub_albums ON child.parent_album_id = sub_albums.id
		)

		SELECT * FROM media WHERE media.album_id IN (
			SELECT id FROM sub_albums
		) AND media.id IN (
			SELECT media_id FROM media_urls WHERE media_urls.media_id = media.id
		) LIMIT 1
	`, obj.ID).Scan(&media).Error

	if err != nil {
		return nil, err
	}

	return &media, nil
}

func (r *albumResolver) SubAlbums(ctx context.Context, parent *models.Album, filter *models.Filter) ([]*models.Album, error) {

	var albums []*models.Album

	query := r.Database.Where("parent_album_id = ?", parent.ID)
	query = filter.FormatSQL(query)

	if err := query.Find(&albums).Error; err != nil {
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

func (r *albumResolver) Shares(ctx context.Context, album *models.Album) ([]*models.ShareToken, error) {

	var shareTokens []*models.ShareToken
	if err := r.Database.Where("album_id = ?", album.ID).Find(&shareTokens).Error; err != nil {
		return nil, err
	}

	return shareTokens, nil
}

func (r *albumResolver) Path(ctx context.Context, obj *models.Album) ([]*models.Album, error) {

	user := auth.UserFromContext(ctx)
	if user == nil {
		empty := make([]*models.Album, 0)
		return empty, nil
	}

	var album_path []*models.Album

	err := r.Database.Raw(`
		WITH recursive path_albums AS (
			SELECT * FROM albums anchor WHERE anchor.id = ?
			UNION
			SELECT parent.* FROM path_albums child JOIN albums parent ON parent.id = child.parent_album_id
		)
		SELECT * FROM path_albums WHERE id != ? AND owner_id = ?
	`, obj.ID, obj.ID, user.ID).Scan(&album_path).Error

	if err != nil {
		return nil, err
	}

	return album_path, nil
}
