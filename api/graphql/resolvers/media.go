package resolvers

import (
	"context"

	api "github.com/photoview/photoview/api/graphql"
	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/dataloader"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/pkg/errors"
	"gorm.io/gorm/clause"
)

func (r *queryResolver) MyMedia(ctx context.Context, order *models.Ordering, paginate *models.Pagination) ([]*models.Media, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, errors.New("unauthorized")
	}

	if err := user.FillAlbums(r.Database); err != nil {
		return nil, err
	}

	userAlbumIDs := make([]int, len(user.Albums))
	for i, album := range user.Albums {
		userAlbumIDs[i] = album.ID
	}

	var media []*models.Media

	query := r.Database.
		Joins("Album").
		Where("albums.id IN (?)", userAlbumIDs).
		Where("media.id IN (?)", r.Database.Model(&models.MediaURL{}).Select("id").Where("media_url.media_id = media.id"))

	query = models.FormatSQL(query, order, paginate)

	if err := query.Scan(&media).Error; err != nil {
		return nil, err
	}

	return media, nil
}

func (r *queryResolver) Media(ctx context.Context, id int, tokenCredentials *models.ShareTokenCredentials) (*models.Media, error) {
	if tokenCredentials != nil {

		shareToken, err := r.ShareToken(ctx, *tokenCredentials)
		if err != nil {
			return nil, err
		}

		if *shareToken.MediaID == id {
			return shareToken.Media, nil
		}
	}

	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	var media models.Media

	err := r.Database.
		Joins("Album").
		Where("media.id = ?", id).
		Where("EXISTS (SELECT * FROM user_albums WHERE user_albums.album_id = media.album_id AND user_albums.user_id = ?)", user.ID).
		Where("media.id IN (?)", r.Database.Model(&models.MediaURL{}).Select("media_id").Where("media_urls.media_id = media.id")).
		First(&media).Error

	if err != nil {
		return nil, errors.Wrap(err, "could not get media by media_id and user_id from database")
	}

	return &media, nil
}

func (r *queryResolver) MediaList(ctx context.Context, ids []int) ([]*models.Media, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	if len(ids) == 0 {
		return nil, errors.New("no ids provided")
	}

	var media []*models.Media
	err := r.Database.Model(&media).
		Joins("LEFT JOIN user_albums ON user_albums.album_id = media.album_id").
		Where("media.id IN ?", ids).
		Where("user_albums.user_id = ?", user.ID).
		Scan(&media).Error

	if err != nil {
		return nil, errors.Wrap(err, "could not get media list by media_id and user_id from database")
	}

	return media, nil
}

type mediaResolver struct {
	*Resolver
}

func (r *Resolver) Media() api.MediaResolver {
	return &mediaResolver{r}
}

func (r *mediaResolver) Shares(ctx context.Context, media *models.Media) ([]*models.ShareToken, error) {
	var shareTokens []*models.ShareToken
	if err := r.Database.Where("media_id = ?", media.ID).Find(&shareTokens).Error; err != nil {
		return nil, errors.Wrapf(err, "get shares for media (%s)", media.Path)
	}

	return shareTokens, nil
}

func (r *mediaResolver) Downloads(ctx context.Context, media *models.Media) ([]*models.MediaDownload, error) {

	var mediaUrls []*models.MediaURL
	if err := r.Database.Where("media_id = ?", media.ID).Find(&mediaUrls).Error; err != nil {
		return nil, errors.Wrapf(err, "get downloads for media (%s)", media.Path)
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
		case url.Purpose == models.VideoThumbnail:
			title = "Video thumbnail"
		case url.Purpose == models.VideoWeb:
			title = "Web optimized video"
		}

		downloads = append(downloads, &models.MediaDownload{
			Title:    title,
			MediaURL: url,
		})
	}

	return downloads, nil
}

func (r *mediaResolver) HighRes(ctx context.Context, media *models.Media) (*models.MediaURL, error) {
	if media.Type != models.MediaTypePhoto {
		return nil, nil
	}

	return dataloader.For(ctx).MediaHighres.Load(media.ID)
}

func (r *mediaResolver) Thumbnail(ctx context.Context, media *models.Media) (*models.MediaURL, error) {
	return dataloader.For(ctx).MediaThumbnail.Load(media.ID)
}

func (r *mediaResolver) VideoWeb(ctx context.Context, media *models.Media) (*models.MediaURL, error) {
	if media.Type != models.MediaTypeVideo {
		return nil, nil
	}

	return dataloader.For(ctx).MediaVideoWeb.Load(media.ID)
}

func (r *mediaResolver) Exif(ctx context.Context, media *models.Media) (*models.MediaEXIF, error) {
	if media.Exif != nil {
		return media.Exif, nil
	}

	var exif models.MediaEXIF
	if err := r.Database.Model(&media).Association("Exif").Find(&exif); err != nil {
		return nil, err
	}

	return &exif, nil
}

func (r *mediaResolver) Favorite(ctx context.Context, media *models.Media) (bool, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return false, auth.ErrUnauthorized
	}

	return dataloader.For(ctx).UserMediaFavorite.Load(&models.UserMediaData{
		UserID:  user.ID,
		MediaID: media.ID,
	})
}

func (r *mutationResolver) FavoriteMedia(ctx context.Context, mediaID int, favorite bool) (*models.Media, error) {

	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	userMediaData := models.UserMediaData{
		UserID:   user.ID,
		MediaID:  mediaID,
		Favorite: favorite,
	}

	if err := r.Database.Clauses(clause.OnConflict{UpdateAll: true}).Create(&userMediaData).Error; err != nil {
		return nil, errors.Wrapf(err, "update user favorite media in database")
	}

	var media models.Media
	if err := r.Database.First(&media, mediaID).Error; err != nil {
		return nil, errors.Wrap(err, "get media from database after favorite update")
	}

	return &media, nil
}

func (r *mediaResolver) Faces(ctx context.Context, media *models.Media) ([]*models.ImageFace, error) {
	if media.Faces != nil {
		return media.Faces, nil
	}

	var faces []*models.ImageFace
	if err := r.Database.Model(&media).Association("Faces").Find(&faces); err != nil {
		return nil, err
	}

	return faces, nil
}
