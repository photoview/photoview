package resolvers

import (
	"context"
	"errors"

	api "github.com/photoview/photoview/api/graphql"
	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
	"gorm.io/gorm"
)

func (r *queryResolver) MyAlbums(ctx context.Context, filter *models.Filter, onlyRoot *bool, showEmpty *bool, onlyWithFavorites *bool) ([]*models.Album, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	if err := user.FillAlbums(r.Database); err != nil {
		return nil, err
	}

	userAlbumIDs := make([]int, len(user.Albums))
	for i, album := range user.Albums {
		userAlbumIDs[i] = album.ID
	}

	query := r.Database.Model(models.Album{}).Where("id IN (?)", userAlbumIDs)

	if onlyRoot != nil && *onlyRoot == true {
		query = query.Where("parent_album_id IS NULL")
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
	if err := query.Scan(&albums).Error; err != nil {
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
	if err := r.Database.First(&album, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("album not found")
		}
		return nil, err
	}

	ownsAlbum, err := user.OwnsAlbum(r.Database, &album)
	if err != nil {
		return nil, err
	}

	if !ownsAlbum {
		return nil, errors.New("forbidden")
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
		SELECT * FROM path_albums WHERE id != ?
	`, obj.ID, obj.ID).Scan(&album_path).Error

	// Make sure to only return albums this user owns
	for i := len(album_path) - 1; i >= 0; i-- {
		album := album_path[i]

		owns, err := user.OwnsAlbum(r.Database, album)
		if err != nil {
			return nil, err
		}

		if !owns {
			album_path = album_path[i+1:]
			break
		}

	}

	if err != nil {
		return nil, err
	}

	return album_path, nil
}
