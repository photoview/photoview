package processing_tasks

import (
	"fmt"
	"path"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/media_encoding"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"gorm.io/gorm"
)

func generateSaveHighResJPEG(tx *gorm.DB, media *models.Media, imageData *media_encoding.EncodeMediaData, highResName string, cacheFs afero.Fs, imagePath string, mediaURL *models.MediaURL) (*models.MediaURL, error) {

	err := imageData.EncodeHighRes(imagePath)
	if err != nil {
		return nil, errors.Wrap(err, "creating high-res cached image")
	}

	photoDimensions, err := media_encoding.GetPhotoDimensions(imagePath)
	if err != nil {
		return nil, err
	}

	fileStats, err := cacheFs.Stat(imagePath)
	if err != nil {
		return nil, errors.Wrap(err, "reading file stats of highres photo")
	}

	if mediaURL == nil {

		mediaURL = &models.MediaURL{
			MediaID:     media.ID,
			MediaName:   highResName,
			Width:       photoDimensions.Width,
			Height:      photoDimensions.Height,
			Purpose:     models.PhotoHighRes,
			ContentType: "image/jpeg",
			FileSize:    fileStats.Size(),
		}

		if err := tx.Create(&mediaURL).Error; err != nil {
			return nil, errors.Wrapf(err, "could not insert highres media url (%d, %s)", media.ID, highResName)
		}
	} else {
		mediaURL.Width = photoDimensions.Width
		mediaURL.Height = photoDimensions.Height
		mediaURL.FileSize = fileStats.Size()

		if err := tx.Save(&mediaURL).Error; err != nil {
			return nil, errors.Wrapf(err, "could not update media url after side car changes (%d, %s)", media.ID, highResName)
		}
	}

	fmt.Printf("Generated high-res: %s (%dx%d)\n", mediaURL.Media.Path, mediaURL.Width, mediaURL.Height)

	return mediaURL, nil
}

func generateSaveThumbnailJPEG(tx *gorm.DB, media *models.Media, cacheFs afero.Fs, thumbnailName string, photoCachePath string, baseImagePath string, mediaURL *models.MediaURL) (*models.MediaURL, error) {
	thumbOutputPath := path.Join(photoCachePath, thumbnailName)

	thumbSize, err := media_encoding.EncodeThumbnail(tx, baseImagePath, thumbOutputPath)
	if err != nil {
		return nil, errors.Wrap(err, "could not create thumbnail cached image")
	}

	fileStats, err := cacheFs.Stat(thumbOutputPath)
	if err != nil {
		return nil, errors.Wrap(err, "reading file stats of thumbnail photo")
	}

	if mediaURL == nil {

		mediaURL = &models.MediaURL{
			MediaID:     media.ID,
			MediaName:   thumbnailName,
			Width:       thumbSize.Width,
			Height:      thumbSize.Height,
			Purpose:     models.PhotoThumbnail,
			ContentType: "image/jpeg",
			FileSize:    fileStats.Size(),
		}

		if err := tx.Create(&mediaURL).Error; err != nil {
			return nil, errors.Wrapf(err, "could not insert thumbnail media url (%d, %s)", media.ID, thumbnailName)
		}
	} else {
		mediaURL.Width = thumbSize.Width
		mediaURL.Height = thumbSize.Height
		mediaURL.FileSize = fileStats.Size()

		if err := tx.Save(&mediaURL).Error; err != nil {
			return nil, errors.Wrapf(err, "could not update media url after side car changes (%d, %s)", media.ID, thumbnailName)
		}
	}

	fmt.Printf("Generated thumbnails: %s (%dx%d)\n", thumbOutputPath, thumbSize.Width, thumbSize.Height)

	return mediaURL, nil
}
