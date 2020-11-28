package resolvers

import (
	"context"

	"github.com/pkg/errors"
	api "github.com/viktorstrate/photoview/api/graphql"
	"github.com/viktorstrate/photoview/api/graphql/auth"
	"github.com/viktorstrate/photoview/api/graphql/models"
	"github.com/viktorstrate/photoview/api/scanner"
	"gorm.io/gorm"
)

func (r *queryResolver) MyMedia(ctx context.Context, filter *models.Filter) ([]*models.Media, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, errors.New("unauthorized")
	}

	filterSQL, err := filter.FormatSQL("media")
	if err != nil {
		return nil, err
	}

	var media []*models.Media
	err = r.Database.Raw(`
		SELECT media.* FROM media, album
		WHERE media.album_id = albums.id AND albums.owner_id = ?
		AND media.media_id IN (
			SELECT media_id FROM media_url WHERE media_url.media_id = media.id
		)
	`+filterSQL, user.ID).Scan(&media).Error
	if err != nil {
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
	err := r.Database.Raw(`
		SELECT media.* FROM media
		JOIN albums ON media.album_id = albums.id
		WHERE media.media_id = ? AND album.owner_id = ?
		AND media.media_id IN (
			SELECT media_id FROM media_url WHERE media_url.media_id = media.media_id
		)
	`, id, user.ID).Scan(&media).Error

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
	// err := r.Database.
	// 	Select("media.*").
	// 	Joins("Album").
	// 	Where("media.id IN ?", ids).
	// 	Where("album.owner_id = ?", user.ID).
	// 	Where("media.id IN (?)", r.Database.Model(&models.MediaURL{}).Select("media_id").Where("media_url.media_id = media.id")).
	// 	Scan(&media).Error

	err := r.Database.Raw(`
		SELECT media.* FROM media
		JOIN albums AS album ON media.album_id = album.id
		WHERE media.media_id IN ? AND album.owner_id = ?
		AND media.media_id IN (
			SELECT media_id FROM media_url WHERE media_url.media_id = media.media_id
		)
	`, ids, user.ID).Error

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
		Where("media_url = ?", media.ID).
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
		Where("media_url = ?", media.ID).
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
		Where("media_url = ?", media.ID).
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

// func (r *mediaResolver) Album(ctx context.Context, media *models.Media) (*models.Album, error) {

// 	r.Database.Model(&media).Joins("Album")

// 	row := r.Database.QueryRow("SELECT album.* from media JOIN album ON media.album_id = album.album_id WHERE media_id = ?", media.MediaID)
// 	return models.NewAlbumFromRow(row)
// }

// func (r *mediaResolver) Exif(ctx context.Context, obj *models.Media) (*models.MediaEXIF, error) {
// 	row := r.Database.QueryRow("SELECT media_exif.* FROM media NATURAL JOIN media_exif WHERE media.media_id = ?", obj.MediaID)

// 	exif, err := models.NewMediaExifFromRow(row)
// 	if err != nil {
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			return nil, nil
// 		} else {
// 			return nil, errors.Wrapf(err, "could not get exif of media from database")
// 		}
// 	}

// 	return exif, nil
// }

// func (r *mediaResolver) VideoMetadata(ctx context.Context, obj *models.Media) (*models.VideoMetadata, error) {
// 	row := r.Database.QueryRow("SELECT video_metadata.* FROM media JOIN video_metadata ON media.video_metadata_id = video_metadata.metadata_id WHERE media.media_id = ?", obj.MediaID)

// 	metadata, err := models.NewVideoMetadataFromRow(row)
// 	if err != nil {
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			return nil, nil
// 		} else {
// 			return nil, errors.Wrapf(err, "could not get video metadata of media from database")
// 		}
// 	}

// 	return metadata, nil
// }

func (r *mutationResolver) FavoriteMedia(ctx context.Context, mediaID int, favorite bool) (*models.Media, error) {

	user := auth.UserFromContext(ctx)

	var media models.Media

	if err := r.Database.Joins("Album").Where("Album.owner_id = ?", user.ID).First(&media, mediaID).Error; err != nil {
		return nil, err
	}

	media.Favorite = favorite

	if err := r.Database.Save(&media).Error; err != nil {
		return nil, errors.Wrap(err, "failed to update media favorite on database")
	}

	return &media, nil
}
