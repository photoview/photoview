package processing_tasks

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/media_encoding"
	"github.com/photoview/photoview/api/scanner/media_encoding/media_utils"
	"github.com/photoview/photoview/api/scanner/scanner_task"
	"github.com/pkg/errors"

	// Image decoders
	_ "image/gif"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

type ProcessPhotoTask struct {
	scanner_task.ScannerTaskBase
}

func (t ProcessPhotoTask) ProcessMedia(ctx scanner_task.TaskContext, mediaData *media_encoding.EncodeMediaData, mediaCachePath string) ([]*models.MediaURL, error) {
	if mediaData.Media.Type != models.MediaTypePhoto {
		return []*models.MediaURL{}, nil
	}

	updatedURLs := make([]*models.MediaURL, 0)
	photo := mediaData.Media

	log.Printf("Processing photo: %s\n", photo.Path)

	photoURLFromDB := makePhotoURLChecker(ctx.GetDB(), photo.ID)

	// original photo url
	origURL, err := photoURLFromDB(models.MediaOriginal)
	if err != nil {
		return []*models.MediaURL{}, err
	}

	// Thumbnail
	thumbURL, err := photoURLFromDB(models.PhotoThumbnail)
	if err != nil {
		return []*models.MediaURL{}, errors.Wrap(err, "error processing photo thumbnail")
	}

	// Highres
	highResURL, err := photoURLFromDB(models.PhotoHighRes)
	if err != nil {
		return []*models.MediaURL{}, errors.Wrap(err, "error processing photo highres")
	}

	var photoDimensions *media_utils.PhotoDimensions
	var baseImagePath string = photo.Path

	// Generate high res jpeg
	if highResURL == nil {

		contentType, err := mediaData.ContentType()
		if err != nil {
			return []*models.MediaURL{}, err
		}

		if !contentType.IsWebCompatible() {
			highresName := generateUniqueMediaNamePrefixed("highres", photo.Path, ".jpg")
			baseImagePath = path.Join(mediaCachePath, highresName)

			highRes, err := generateSaveHighResJPEG(ctx.GetDB(), photo, mediaData, highresName, baseImagePath, nil)
			if err != nil {
				return []*models.MediaURL{}, err
			}

			updatedURLs = append(updatedURLs, highRes)
		}
	} else {
		// Verify that highres photo still exists in cache
		baseImagePath = path.Join(mediaCachePath, highResURL.MediaName)

		if _, err := os.Stat(baseImagePath); os.IsNotExist(err) {
			fmt.Printf("High-res photo found in database but not in cache, re-encoding photo to cache: %s\n", highResURL.MediaName)
			updatedURLs = append(updatedURLs, highResURL)

			err = mediaData.EncodeHighRes(baseImagePath)
			if err != nil {
				return []*models.MediaURL{}, errors.Wrap(err, "creating high-res cached image")
			}
		}
	}

	// Save original photo to database
	if origURL == nil {

		// Make sure photo dimensions is set
		if photoDimensions == nil {
			photoDimensions, err = media_utils.GetPhotoDimensions(baseImagePath)
			if err != nil {
				return []*models.MediaURL{}, err
			}
		}

		original, err := saveOriginalPhotoToDB(ctx.GetDB(), photo, mediaData, photoDimensions)
		if err != nil {
			return []*models.MediaURL{}, errors.Wrap(err, "saving original photo to database")
		}

		updatedURLs = append(updatedURLs, original)
	}

	// Save thumbnail to cache
	if thumbURL == nil {
		thumbnailName := generateUniqueMediaNamePrefixed("thumbnail", photo.Path, ".jpg")
		thumbnail, err := generateSaveThumbnailJPEG(ctx.GetDB(), photo, thumbnailName, mediaCachePath, baseImagePath, nil)
		if err != nil {
			return []*models.MediaURL{}, err
		}

		updatedURLs = append(updatedURLs, thumbnail)
	} else {
		// Verify that thumbnail photo still exists in cache
		thumbPath := path.Join(mediaCachePath, thumbURL.MediaName)

		if _, err := os.Stat(thumbPath); os.IsNotExist(err) {
			updatedURLs = append(updatedURLs, thumbURL)
			fmt.Printf("Thumbnail photo found in database but not in cache, re-encoding photo to cache: %s\n", thumbURL.MediaName)

			_, err := media_encoding.EncodeThumbnail(ctx.GetDB(), baseImagePath, thumbPath)
			if err != nil {
				return []*models.MediaURL{}, errors.Wrap(err, "could not create thumbnail cached image")
			}
		}
	}

	return updatedURLs, nil
}
