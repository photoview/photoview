package processing_tasks

import (
	"os"
	"path"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/media_encoding"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func generateSaveHighResJPEG(tx *gorm.DB, media *models.Media, imageData *media_encoding.EncodeMediaData, highResName string, imagePath string, mediaURL *models.MediaURL) (*models.MediaURL, error) {

	err := imageData.EncodeHighRes(imagePath)
	if err != nil {
		return nil, errors.Wrap(err, "creating high-res cached image")
	}

	photoDimensions, err := media_encoding.GetPhotoDimensions(imagePath)
	if err != nil {
		return nil, err
	}

	fileStats, err := os.Stat(imagePath)
	if err != nil {
		return nil, errors.Wrap(err, "reading file stats of highres photo")
	}

	if mediaURL == nil {
		mediaURL = &models.MediaURL{
			MediaID: media.ID,
			Purpose: models.PhotoHighRes,
		}
	}
	mediaURL.MediaName = highResName
	mediaURL.Width = photoDimensions.Width
	mediaURL.Height = photoDimensions.Height
	mediaURL.ContentType = "image/jpeg"
	mediaURL.FileSize = fileStats.Size()

	if err := models.UpsertMediaURL(tx, mediaURL); err != nil {
		return nil, errors.Wrapf(err, "could not save highres media url (%d, %s)", media.ID, highResName)
	}

	return mediaURL, nil
}

func generateSaveThumbnailJPEG(tx *gorm.DB, media *models.Media, thumbnailName string, photoCachePath string, baseImagePath string, mediaURL *models.MediaURL) (*models.MediaURL, error) {
	thumbOutputPath := path.Join(photoCachePath, thumbnailName)

	thumbSize, err := media_encoding.EncodeThumbnail(tx, baseImagePath, thumbOutputPath)
	if err != nil {
		return nil, errors.Wrap(err, "could not create thumbnail cached image")
	}

	fileStats, err := os.Stat(thumbOutputPath)
	if err != nil {
		return nil, errors.Wrap(err, "reading file stats of thumbnail photo")
	}

	if mediaURL == nil {
		mediaURL = &models.MediaURL{
			MediaID: media.ID,
			Purpose: models.PhotoThumbnail,
		}
	}
	mediaURL.MediaName = thumbnailName
	mediaURL.Width = thumbSize.Width
	mediaURL.Height = thumbSize.Height
	mediaURL.ContentType = "image/jpeg"
	mediaURL.FileSize = fileStats.Size()

	if err := models.UpsertMediaURL(tx, mediaURL); err != nil {
		return nil, errors.Wrapf(err, "could not save thumbnail media url (%d, %s)", media.ID, thumbnailName)
	}

	return mediaURL, nil
}
