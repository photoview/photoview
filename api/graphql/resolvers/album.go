package resolvers

import (
	"context"

	"fmt"

	api "github.com/photoview/photoview/api/graphql"
	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func (r *queryResolver) MyAlbums(ctx context.Context, order *models.Ordering, paginate *models.Pagination, onlyRoot *bool, showEmpty *bool, onlyWithFavorites *bool) ([]*models.Album, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	if err := user.FillAlbums(r.Database); err != nil {
		return nil, err
	}

	if len(user.Albums) == 0 {
		return nil, nil
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
			favoritesSubquery := r.Database.
				Model(&models.UserMediaData{UserID: user.ID}).
				Where("user_media_data.media_id = media.id").
				Where("user_media_data.favorite = true")

			subQuery = subQuery.Where("EXISTS (?)", favoritesSubquery)
		}

		query = query.Where("EXISTS (?)", subQuery)
	}

	query = models.FormatSQL(query, order, paginate)

	var albums []*models.Album
	if err := query.Find(&albums).Error; err != nil {
		return nil, err
	}

	return albums, nil
}

func (r *queryResolver) Album(ctx context.Context, id int, tokenCredentials *models.ShareTokenCredentials) (*models.Album, error) {
	if tokenCredentials != nil {

		shareToken, err := r.ShareToken(ctx, *tokenCredentials)
		if err != nil {
			return nil, err
		}

		if shareToken.Album != nil {
			if *shareToken.AlbumID == id {
				return shareToken.Album, nil
			}

			subAlbum, err := shareToken.Album.GetChildren(r.Database, func(query *gorm.DB) *gorm.DB { return query.Where("sub_albums.id = ?", id) })
			if err != nil {
				return nil, errors.Wrapf(err, "find sub album of share token (%s)", tokenCredentials.Token)
			}

			if len(subAlbum) > 0 {
				return subAlbum[0], nil
			}
		}
	}

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

func (r *albumResolver) Media(ctx context.Context, album *models.Album, order *models.Ordering, paginate *models.Pagination, onlyFavorites *bool) ([]*models.Media, error) {

	query := r.Database.
		Where("media.album_id = ?", album.ID).
		Where("media.id IN (?)", r.Database.Model(&models.MediaURL{}).Select("media_urls.media_id").Where("media_urls.media_id = media.id"))

	if onlyFavorites != nil && *onlyFavorites == true {
		user := auth.UserFromContext(ctx)
		if user == nil {
			return nil, errors.New("cannot get favorite media without being authorized")
		}

		favoriteQuery := r.Database.Model(&models.UserMediaData{
			UserID: user.ID,
		}).Where("user_media_data.media_id = media.id").Where("user_media_data.favorite = true")

		query = query.Where("EXISTS (?)", favoriteQuery)
	}

	query = models.FormatSQL(query, order, paginate)

	var media []*models.Media
	if err := query.Find(&media).Error; err != nil {
		return nil, err
	}

	return media, nil
}

func (r *albumResolver) Thumbnail(ctx context.Context, obj *models.Album) (*models.Media, error) {

	var media models.Media

	fmt.Print(obj.CoverID)

	if obj.CoverID == -1 {
		if err := r.Database.Raw(`
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
		`, obj.ID).Find(&media).Error; err != nil {
			return nil, err
		}
	} else {
		if err := r.Database.Where("id = ?", obj.CoverID).Find(&media).Error; err != nil {
			return nil, err
		}
	}

	// err := r.Database.Raw(`
	// 	WITH recursive sub_albums AS (
	// 		SELECT * FROM albums AS root WHERE id = ?
	// 		UNION ALL
	// 		SELECT child.* FROM albums AS child JOIN sub_albums ON child.parent_album_id = sub_albums.id
	// 	)
	//
	// 	SELECT * FROM media WHERE media.album_id IN (
	// 		SELECT id FROM sub_albums
	// 	) AND media.id IN (
	// 		SELECT media_id FROM media_urls WHERE media_urls.media_id = media.id
	// 	) LIMIT 1
	// `, obj.ID).Find(&media).Error

	return &media, nil
}

func (r *albumResolver) SubAlbums(ctx context.Context, parent *models.Album, order *models.Ordering, paginate *models.Pagination) ([]*models.Album, error) {

	var albums []*models.Album

	query := r.Database.Where("parent_album_id = ?", parent.ID)
	query = models.FormatSQL(query, order, paginate)

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

func (r *mutationResolver) SetAlbumCoverID(ctx context.Context, albumID int, coverID *int) (*models.Album, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, errors.New("unauthorized")
	}

	var album models.Album
	if err := r.Database.Find(&album, albumID).Error; err != nil {
		return nil, err
	}

	//
	// var album models.Album
	//
	ownsAlbum, err := user.OwnsAlbum(r.Database, &album)
	if err != nil {
		return nil, err
	}

	if !ownsAlbum {
		return nil, errors.New("forbidden")
	}

	if err := r.Database.Model(&album).Update("cover_id", coverID).Error; err != nil {
		return nil, err
	}

	// var faceGroup models.FaceGroup
	// if err := db.Where("id = ?", faceGroupID).Find(&faceGroup).Error; err != nil {
	// 	return nil, err
	// }
	//
	// return &faceGroup, nil
	//
	// if err := r.Database.Model(faceGroup).Update("label", label).Error; err != nil {
	// 	return nil, err
	// }

	return &album, nil
}
