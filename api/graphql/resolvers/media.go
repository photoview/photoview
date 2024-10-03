package resolvers

import (
	"context"
	"strings"

	"github.com/photoview/photoview/api/dataloader"
	api "github.com/photoview/photoview/api/graphql"
	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/graphql/models/actions"
	"github.com/photoview/photoview/api/scanner/face_detection"
	"github.com/pkg/errors"
)

func (r *queryResolver) MyMedia(ctx context.Context, order *models.Ordering, paginate *models.Pagination) ([]*models.Media,
	error) {

	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, errors.New("unauthorized")
	}

	return actions.MyMedia(r.DB(ctx), user, order, paginate)
}

func (r *queryResolver) Media(ctx context.Context, id int, tokenCredentials *models.ShareTokenCredentials) (*models.Media,
	error) {

	db := r.DB(ctx)
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

	err := db.
		Joins("Album").
		Where("media.id = ?", id).
		Where("EXISTS (SELECT * FROM user_albums WHERE user_albums.album_id = media.album_id AND user_albums.user_id = ?)",
			user.ID).
		Where("media.id IN (?)", db.Model(&models.MediaURL{}).Select("media_id").Where("media_urls.media_id = media.id")).
		First(&media).Error

	if err != nil {
		return nil, errors.Wrap(err, "could not get media by media_id and user_id from database")
	}

	return &media, nil
}

func (r *queryResolver) MediaList(ctx context.Context, ids []int) ([]*models.Media, error) {
	db := r.DB(ctx)
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	if len(ids) == 0 {
		return nil, errors.New("no ids provided")
	}

	var media []*models.Media
	err := db.Model(&media).
		Joins("LEFT JOIN user_albums ON user_albums.album_id = media.album_id").
		Where("media.id IN ?", ids).
		Where("user_albums.user_id = ?", user.ID).
		Find(&media).Error

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

func (r *mediaResolver) Type(ctx context.Context, media *models.Media) (models.MediaType, error) {
	formattedType := models.MediaType(strings.Title(string(media.Type)))
	return formattedType, nil
}

func (r *mediaResolver) Album(ctx context.Context, obj *models.Media) (*models.Album, error) {
	var album models.Album
	err := r.DB(ctx).Find(&album, obj.AlbumID).Error
	if err != nil {
		return nil, err
	}
	return &album, nil
}

func (r *mediaResolver) Shares(ctx context.Context, media *models.Media) ([]*models.ShareToken, error) {
	var shareTokens []*models.ShareToken
	if err := r.DB(ctx).Where("media_id = ?", media.ID).Find(&shareTokens).Error; err != nil {
		return nil, errors.Wrapf(err, "get shares for media (%s)", media.Path)
	}

	return shareTokens, nil
}

func (r *mediaResolver) Downloads(ctx context.Context, media *models.Media) ([]*models.MediaDownload, error) {

	var mediaUrls []*models.MediaURL
	if err := r.DB(ctx).Where("media_id = ?", media.ID).Find(&mediaUrls).Error; err != nil {
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
	if err := r.DB(ctx).Model(&media).Association("Exif").Find(&exif); err != nil {
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

	return user.FavoriteMedia(r.DB(ctx), mediaID, favorite)
}

func (r *mediaResolver) Faces(ctx context.Context, media *models.Media) ([]*models.ImageFace, error) {
	if face_detection.GlobalFaceDetector == nil {
		return []*models.ImageFace{}, nil
	}

	if media.Faces != nil {
		return media.Faces, nil
	}

	var faces []*models.ImageFace
	if err := r.DB(ctx).Model(&media).Association("Faces").Find(&faces); err != nil {
		return nil, err
	}

	return faces, nil
}
