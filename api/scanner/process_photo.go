package scanner

import (
	"fmt"
	"log"
	"os"
	"path"
	"strconv"

	"github.com/pkg/errors"
	"github.com/viktorstrate/photoview/api/graphql/models"
	"github.com/viktorstrate/photoview/api/utils"
	"gorm.io/gorm"

	// Image decoders
	_ "image/gif"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

// Higher order function used to check if MediaURL for a given MediaPurpose exists
func makePhotoURLChecker(tx *gorm.DB, mediaID int) func(purpose models.MediaPurpose) (*models.MediaURL, error) {
	return func(purpose models.MediaPurpose) (*models.MediaURL, error) {
		var mediaURL []*models.MediaURL

		result := tx.Where("purpose = ?", purpose).Where("media_id = ?", mediaID).Find(&mediaURL)

		if result.Error != nil {
			return nil, result.Error
		}

		if result.RowsAffected > 0 {
			return mediaURL[0], nil
		}

		return nil, nil
	}
}

func ProcessMedia(tx *gorm.DB, media *models.Media) (bool, error) {
	imageData := EncodeMediaData{
		media: media,
	}

	contentType, err := imageData.ContentType()
	if err != nil {
		return false, errors.Wrapf(err, "get content-type of media (%s)", media.Path)
	}

	// Make sure media cache directory exists
	mediaCachePath, err := makeMediaCacheDir(media)
	if err != nil {
		return false, errors.Wrap(err, "cache directory error")
	}

	if contentType.isVideo() {
		return processVideo(tx, &imageData, mediaCachePath)
	} else {
		return processPhoto(tx, &imageData, mediaCachePath)
	}
}

func processPhoto(tx *gorm.DB, imageData *EncodeMediaData, photoCachePath *string) (bool, error) {

	photo := imageData.media

	log.Printf("Processing photo: %s\n", photo.Path)

	didProcess := false

	photoURLFromDB := makePhotoURLChecker(tx, photo.ID)

	// original photo url
	origURL, err := photoURLFromDB(models.MediaOriginal)
	if err != nil {
		return false, err
	}

	// Thumbnail
	thumbURL, err := photoURLFromDB(models.PhotoThumbnail)
	if err != nil {
		return false, errors.Wrap(err, "error processing photo thumbnail")
	}

	// Highres
	highResURL, err := photoURLFromDB(models.PhotoHighRes)
	if err != nil {
		return false, errors.Wrap(err, "error processing photo highres")
	}

	var photoDimensions *PhotoDimensions
	var baseImagePath string = photo.Path

	mediaType, err := getMediaType(photo.Path)
	if err != nil {
		return false, errors.Wrap(err, "could determine if media was photo or video")
	}

	if mediaType.isRaw() {
		err = processRawSideCar(tx, imageData, highResURL, thumbURL, photoCachePath)
		if err != nil {
			return false, err
		}

		counterpartFile := scanForCompressedCounterpartFile(photo.Path)
		if counterpartFile != nil {
			photo.CounterpartPath = counterpartFile
		}
	}

	// Generate high res jpeg
	if highResURL == nil {

		contentType, err := imageData.ContentType()
		if err != nil {
			return false, err
		}

		if !contentType.isWebCompatible() {
			didProcess = true

			highres_name := fmt.Sprintf("highres_%s_%s", path.Base(photo.Path), utils.GenerateToken())
			highres_name = models.SanitizeMediaName(highres_name)
			highres_name = highres_name + ".jpg"

			baseImagePath = path.Join(*photoCachePath, highres_name)

			err = generateSaveHighResJPEG(tx, photo, imageData, highres_name, baseImagePath, nil)
			if err != nil {
				return false, err
			}
		}
	} else {
		// Verify that highres photo still exists in cache
		baseImagePath = path.Join(*photoCachePath, highResURL.MediaName)

		if _, err := os.Stat(baseImagePath); os.IsNotExist(err) {
			fmt.Printf("High-res photo found in database but not in cache, re-encoding photo to cache: %s\n", highResURL.MediaName)
			didProcess = true

			err = imageData.EncodeHighRes(tx, baseImagePath)
			if err != nil {
				return false, errors.Wrap(err, "creating high-res cached image")
			}
		}
	}

	// Save original photo to database
	if origURL == nil {
		didProcess = true

		// Make sure photo dimensions is set
		if photoDimensions == nil {
			photoDimensions, err = GetPhotoDimensions(baseImagePath)
			if err != nil {
				return false, err
			}
		}

		if err = saveOriginalPhotoToDB(tx, photo, imageData, photoDimensions); err != nil {
			return false, errors.Wrap(err, "saving original photo to database")
		}
	}

	// Save thumbnail to cache
	if thumbURL == nil {
		didProcess = true

		thumbnail_name := fmt.Sprintf("thumbnail_%s_%s", path.Base(photo.Path), utils.GenerateToken())
		thumbnail_name = models.SanitizeMediaName(thumbnail_name)
		thumbnail_name = thumbnail_name + ".jpg"

		err = generateSaveThumbnailJPEG(tx, photo, thumbnail_name, photoCachePath, baseImagePath, nil)
		if err != nil {
			return false, err
		}
	} else {
		// Verify that thumbnail photo still exists in cache
		thumbPath := path.Join(*photoCachePath, thumbURL.MediaName)

		if _, err := os.Stat(thumbPath); os.IsNotExist(err) {
			didProcess = true
			fmt.Printf("Thumbnail photo found in database but not in cache, re-encoding photo to cache: %s\n", thumbURL.MediaName)

			_, err := EncodeThumbnail(baseImagePath, thumbPath)
			if err != nil {
				return false, errors.Wrap(err, "could not create thumbnail cached image")
			}
		}
	}

	return didProcess, nil
}

func makeMediaCacheDir(media *models.Media) (*string, error) {

	// Make root cache dir if not exists
	if _, err := os.Stat(PhotoCache()); os.IsNotExist(err) {
		if err := os.Mkdir(PhotoCache(), os.ModePerm); err != nil {
			return nil, errors.Wrap(err, "could not make root image cache directory")
		}
	}

	// Make album cache dir if not exists
	albumCachePath := path.Join(PhotoCache(), strconv.Itoa(int(media.AlbumID)))
	if _, err := os.Stat(albumCachePath); os.IsNotExist(err) {
		if err := os.Mkdir(albumCachePath, os.ModePerm); err != nil {
			return nil, errors.Wrap(err, "could not make album image cache directory")
		}
	}

	// Make photo cache dir if not exists
	photoCachePath := path.Join(albumCachePath, strconv.Itoa(int(media.ID)))
	if _, err := os.Stat(photoCachePath); os.IsNotExist(err) {
		if err := os.Mkdir(photoCachePath, os.ModePerm); err != nil {
			return nil, errors.Wrap(err, "could not make photo image cache directory")
		}
	}

	return &photoCachePath, nil
}

func saveOriginalPhotoToDB(tx *gorm.DB, photo *models.Media, imageData *EncodeMediaData, photoDimensions *PhotoDimensions) error {
	photoName := path.Base(photo.Path)
	photoBaseName := photoName[0 : len(photoName)-len(path.Ext(photoName))]
	photoBaseExt := path.Ext(photoName)

	original_image_name := fmt.Sprintf("%s_%s", photoBaseName, utils.GenerateToken())
	original_image_name = models.SanitizeMediaName(original_image_name) + photoBaseExt

	contentType, err := imageData.ContentType()
	if err != nil {
		return err
	}

	fileStats, err := os.Stat(photo.Path)
	if err != nil {
		return errors.Wrap(err, "reading file stats of original photo")
	}

	mediaURL := models.MediaURL{
		Media:       *photo,
		MediaName:   original_image_name,
		Width:       photoDimensions.Width,
		Height:      photoDimensions.Height,
		Purpose:     models.MediaOriginal,
		ContentType: string(*contentType),
		FileSize:    fileStats.Size(),
	}

	if err := tx.Create(&mediaURL).Error; err != nil {
		return errors.Wrapf(err, "inserting original photo url: %d, %s", photo.ID, photoName)
	}

	return nil
}

func generateSaveHighResJPEG(tx *gorm.DB, media *models.Media, imageData *EncodeMediaData, highres_name string, imagePath string, mediaURL *models.MediaURL) error {

	err := imageData.EncodeHighRes(tx, imagePath)
	if err != nil {
		return errors.Wrap(err, "creating high-res cached image")
	}

	photoDimensions, err := GetPhotoDimensions(imagePath)
	if err != nil {
		return err
	}

	fileStats, err := os.Stat(imagePath)
	if err != nil {
		return errors.Wrap(err, "reading file stats of highres photo")
	}

	if mediaURL == nil {

		mediaURL = &models.MediaURL{
			MediaID:     media.ID,
			MediaName:   highres_name,
			Width:       photoDimensions.Width,
			Height:      photoDimensions.Height,
			Purpose:     models.PhotoHighRes,
			ContentType: "image/jpeg",
			FileSize:    fileStats.Size(),
		}

		if err := tx.Create(&mediaURL).Error; err != nil {
			return errors.Wrapf(err, "could not insert highres media url (%d, %s)", media.ID, highres_name)
		}
	} else {
		mediaURL.Width = photoDimensions.Width
		mediaURL.Height = photoDimensions.Height
		mediaURL.FileSize = fileStats.Size()

		if err := tx.Save(&mediaURL).Error; err != nil {
			return errors.Wrapf(err, "could not update media url after side car changes (%d, %s)", media.ID, highres_name)
		}
	}

	return nil
}

func generateSaveThumbnailJPEG(tx *gorm.DB, media *models.Media, thumbnail_name string, photoCachePath *string, baseImagePath string, mediaURL *models.MediaURL) error {
	thumbOutputPath := path.Join(*photoCachePath, thumbnail_name)

	thumbSize, err := EncodeThumbnail(baseImagePath, thumbOutputPath)
	if err != nil {
		return errors.Wrap(err, "could not create thumbnail cached image")
	}

	fileStats, err := os.Stat(thumbOutputPath)
	if err != nil {
		return errors.Wrap(err, "reading file stats of thumbnail photo")
	}

	if mediaURL == nil {

		mediaURL = &models.MediaURL{
			MediaID:     media.ID,
			MediaName:   thumbnail_name,
			Width:       thumbSize.Width,
			Height:      thumbSize.Height,
			Purpose:     models.PhotoThumbnail,
			ContentType: "image/jpeg",
			FileSize:    fileStats.Size(),
		}

		if err := tx.Create(&mediaURL).Error; err != nil {
			return errors.Wrapf(err, "could not insert thumbnail media url (%d, %s)", media.ID, thumbnail_name)
		}
	} else {
		mediaURL.Width = thumbSize.Width
		mediaURL.Height = thumbSize.Height
		mediaURL.FileSize = fileStats.Size()

		if err := tx.Save(&mediaURL).Error; err != nil {
			return errors.Wrapf(err, "could not update media url after side car changes (%d, %s)", media.ID, thumbnail_name)
		}
	}

	return nil
}

func processRawSideCar(tx *gorm.DB, imageData *EncodeMediaData, highResURL *models.MediaURL, thumbURL *models.MediaURL, photoCachePath *string) error {
	photo := imageData.media
	sideCarFileHasChanged := false
	var currentFileHash *string
	currentSideCarPath := scanForSideCarFile(photo.Path)

	if currentSideCarPath != nil {
		currentFileHash = hashSideCarFile(currentSideCarPath)
		if photo.SideCarHash == nil || *photo.SideCarHash != *currentFileHash {
			sideCarFileHasChanged = true
		}
	} else if photo.SideCarPath != nil { // sidecar has been deleted since last scan
		sideCarFileHasChanged = true
	}

	if sideCarFileHasChanged {
		fmt.Printf("Detected changed sidecar file for %s recreating JPG's to reflect changes\n", photo.Path)

		// update high res image may be cropped so dimentions and file size can change
		baseImagePath := path.Join(*photoCachePath, highResURL.MediaName) // update base image path for thumbnail
		tempHighResPath := baseImagePath + ".hold"
		os.Rename(baseImagePath, tempHighResPath)
		err := generateSaveHighResJPEG(tx, photo, imageData, highResURL.MediaName, baseImagePath, highResURL)
		if err != nil {
			os.Rename(tempHighResPath, baseImagePath)
			return errors.Wrap(err, "recreating high-res cached image")
		}
		os.Remove(tempHighResPath)

		// update thumbnail image may be cropped so dimentions and file size can change
		thumbPath := path.Join(*photoCachePath, thumbURL.MediaName)
		tempThumbPath := thumbPath + ".hold" // hold onto the original image incase for some reason we fail to recreate one with the new settings
		os.Rename(thumbPath, tempThumbPath)
		err = generateSaveThumbnailJPEG(tx, photo, thumbURL.MediaName, photoCachePath, baseImagePath, thumbURL)
		if err != nil {
			os.Rename(tempThumbPath, thumbPath)
			return errors.Wrap(err, "recreating thumbnail cached image")
		}
		os.Remove(tempThumbPath)

		photo.SideCarHash = currentFileHash
		photo.SideCarPath = currentSideCarPath

		// save new side car hash
		if err := tx.Save(&photo).Error; err != nil {
			return errors.Wrapf(err, "could not update side car hash for media: %s", photo.Path)
		}
	}

	return nil
}
