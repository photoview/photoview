package resolvers

import (
	"context"

	api "github.com/photoview/photoview/api/graphql"
	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (r *queryResolver) MyMedia(ctx context.Context, filter *models.Filter) ([]*models.Media, error) {
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

	query = filter.FormatSQL(query)

	if err := query.Scan(&media).Error; err != nil {
		return nil, err
	}

	return media, nil
}

func (r *queryResolver) Media(ctx context.Context, id int) (*models.Media, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	var media models.Media

	err := r.Database.
		Joins("Album").
		Where("media.id = ?", id).
		Where("Album.owner_id = ?", user.ID).
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
	err := r.Database.
		Select("media.*").
		Joins("Album").
		Where("media.id IN ?", ids).
		Where("album.owner_id = ?", user.ID).
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
	var url models.MediaURL
	err := r.Database.
		Where("media_id = ?", media.ID).
		Where("purpose = ? OR (purpose = ? AND content_type IN ?)", models.PhotoHighRes, models.MediaOriginal, scanner.WebMimetypes).
		First(&url).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		} else {
			return nil, errors.Wrapf(err, "could not query high-res (%s)", media.Path)
		}
	}

	return &url, nil
}

func (r *mediaResolver) Thumbnail(ctx context.Context, media *models.Media) (*models.MediaURL, error) {
	var url models.MediaURL
	err := r.Database.
		Where("media_id = ?", media.ID).
		Where("purpose = ? OR purpose = ?", models.PhotoThumbnail, models.VideoThumbnail).
		First(&url).Error

	if err != nil {
		return nil, errors.Wrapf(err, "could not query thumbnail (%s)", media.Path)
	}

	return &url, nil
}

func (r *mediaResolver) VideoWeb(ctx context.Context, media *models.Media) (*models.MediaURL, error) {

	var url models.MediaURL
	err := r.Database.
		Where("media_id = ?", media.ID).
		Where("purpose = ?", models.VideoWeb).
		First(&url).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		} else {
			return nil, errors.Wrapf(err, "could not query video web-format url (%s)", media.Path)
		}
	}

	return &url, nil
}

func (r *mediaResolver) Favorite(ctx context.Context, media *models.Media) (bool, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return false, auth.ErrUnauthorized
	}

	userMediaData := models.UserMediaData{
		UserID:   user.ID,
		MediaID:  media.ID,
		Favorite: false,
	}

	if err := r.Database.FirstOrInit(&userMediaData).Error; err != nil {
		return false, errors.Wrapf(err, "get user media data from database (user: %d, media: %d)", user.ID, media.ID)
	}

	return userMediaData.Favorite, nil
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
