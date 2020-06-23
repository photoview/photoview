package scanner

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/viktorstrate/photoview/api/graphql/models"
	"github.com/viktorstrate/photoview/api/utils"

	// Image decoders
	_ "image/gif"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

// Higher order function used to check if PhotoURL for a given PhotoPurpose exists
func makePhotoURLChecker(tx *sql.Tx, photoID int) (func(purpose models.PhotoPurpose) (*models.PhotoURL, error), error) {
	photoURLExistsStmt, err := tx.Prepare("SELECT * FROM photo_url WHERE photo_id = ? AND purpose = ?")
	if err != nil {
		return nil, err
	}

	return func(purpose models.PhotoPurpose) (*models.PhotoURL, error) {
		row := photoURLExistsStmt.QueryRow(photoID, purpose)
		photoURL, err := models.NewPhotoURLFromRow(row)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			return nil, err
		}

		return photoURL, nil
	}, nil
}

func ProcessPhoto(tx *sql.Tx, photo *models.Photo) (bool, error) {

	log.Printf("Processing photo: %s\n", photo.Path)

	didProcess := false

	imageData := EncodeImageData{
		photo: photo,
	}

	photoUrlFromDB, err := makePhotoURLChecker(tx, photo.PhotoID)
	if err != nil {
		return false, err
	}

	// original photo url
	origURL, err := photoUrlFromDB(models.PhotoOriginal)
	if err != nil {
		return false, err
	}

	// Thumbnail
	thumbURL, err := photoUrlFromDB(models.PhotoThumbnail)
	if err != nil {
		return false, errors.Wrap(err, "error processing thumbnail")
	}

	// Highres
	highResURL, err := photoUrlFromDB(models.PhotoHighRes)
	if err != nil {
		return false, errors.Wrap(err, "error processing highres")
	}

	// Make sure photo cache directory exists
	photoCachePath, err := makePhotoCacheDir(photo)
	if err != nil {
		return false, errors.Wrap(err, "cache directory error")
	}

	// Generate high res jpeg
	var photoDimensions *PhotoDimensions
	var baseImagePath string = photo.Path

	if highResURL == nil {

		contentType, err := imageData.ContentType()
		if err != nil {
			return false, err
		}

		if !contentType.isWebCompatible() {
			didProcess = true

			highres_name := fmt.Sprintf("highres_%s_%s", path.Base(photo.Path), utils.GenerateToken())
			highres_name = strings.ReplaceAll(highres_name, ".", "_")
			highres_name = strings.ReplaceAll(highres_name, " ", "_")
			highres_name = highres_name + ".jpg"

			baseImagePath = path.Join(*photoCachePath, highres_name)

			err = imageData.EncodeHighRes(tx, baseImagePath)
			if err != nil {
				return false, errors.Wrap(err, "creating high-res cached image")
			}

			photoDimensions, err = GetPhotoDimensions(baseImagePath)
			if err != nil {
				return false, err
			}

			_, err = tx.Exec("INSERT INTO photo_url (photo_id, photo_name, width, height, purpose, content_type) VALUES (?, ?, ?, ?, ?, ?)",
				photo.PhotoID, highres_name, photoDimensions.Width, photoDimensions.Height, models.PhotoHighRes, "image/jpeg")
			if err != nil {
				log.Printf("Could not insert highres photo url: %d, %s\n", photo.PhotoID, path.Base(photo.Path))
				return false, err
			}
		}
	} else {
		// Verify that highres photo still exists in cache
		baseImagePath = path.Join(*photoCachePath, highResURL.PhotoName)

		if _, err := os.Stat(baseImagePath); os.IsNotExist(err) {
			fmt.Printf("High-res photo found in database but not in cache, re-encoding photo to cache: %s\n", highResURL.PhotoName)
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
		thumbnail_name = strings.ReplaceAll(thumbnail_name, ".", "_")
		thumbnail_name = strings.ReplaceAll(thumbnail_name, " ", "_")
		thumbnail_name = thumbnail_name + ".jpg"

		// thumbnailImage, err := imageData.ThumbnailImage(tx)
		// if err != nil {
		// 	return err
		// }

		thumbOutputPath := path.Join(*photoCachePath, thumbnail_name)

		thumbSize, err := EncodeThumbnail(baseImagePath, thumbOutputPath)
		if err != nil {
			return false, errors.Wrap(err, "could not create thumbnail cached image")
		}

		_, err = tx.Exec("INSERT INTO photo_url (photo_id, photo_name, width, height, purpose, content_type) VALUES (?, ?, ?, ?, ?, ?)", photo.PhotoID, thumbnail_name, thumbSize.Width, thumbSize.Height, models.PhotoThumbnail, "image/jpeg")
		if err != nil {
			return false, err
		}
	} else {
		// Verify that thumbnail photo still exists in cache
		thumbPath := path.Join(*photoCachePath, thumbURL.PhotoName)

		if _, err := os.Stat(thumbPath); os.IsNotExist(err) {
			didProcess = true
			fmt.Printf("Thumbnail photo found in database but not in cache, re-encoding photo to cache: %s\n", thumbURL.PhotoName)

			_, err := EncodeThumbnail(baseImagePath, thumbPath)
			if err != nil {
				return false, errors.Wrap(err, "could not create thumbnail cached image")
			}
		}
	}

	return didProcess, nil
}

func makePhotoCacheDir(photo *models.Photo) (*string, error) {

	// Make root cache dir if not exists
	if _, err := os.Stat(PhotoCache()); os.IsNotExist(err) {
		if err := os.Mkdir(PhotoCache(), os.ModePerm); err != nil {
			log.Println("ERROR: Could not make root image cache directory")
			return nil, err
		}
	}

	// Make album cache dir if not exists
	albumCachePath := path.Join(PhotoCache(), strconv.Itoa(photo.AlbumId))
	if _, err := os.Stat(albumCachePath); os.IsNotExist(err) {
		if err := os.Mkdir(albumCachePath, os.ModePerm); err != nil {
			log.Println("ERROR: Could not make album image cache directory")
			return nil, err
		}
	}

	// Make photo cache dir if not exists
	photoCachePath := path.Join(albumCachePath, strconv.Itoa(photo.PhotoID))
	if _, err := os.Stat(photoCachePath); os.IsNotExist(err) {
		if err := os.Mkdir(photoCachePath, os.ModePerm); err != nil {
			log.Println("ERROR: Could not make photo image cache directory")
			return nil, err
		}
	}

	return &photoCachePath, nil
}

func saveOriginalPhotoToDB(tx *sql.Tx, photo *models.Photo, imageData EncodeImageData, photoDimensions *PhotoDimensions) error {
	photoName := path.Base(photo.Path)
	photoBaseName := photoName[0 : len(photoName)-len(path.Ext(photoName))]
	photoBaseExt := path.Ext(photoName)

	original_image_name := fmt.Sprintf("%s_%s", photoBaseName, utils.GenerateToken())
	original_image_name = strings.ReplaceAll(original_image_name, " ", "_") + photoBaseExt

	contentType, err := imageData.ContentType()
	if err != nil {
		return err
	}

	_, err = tx.Exec("INSERT INTO photo_url (photo_id, photo_name, width, height, purpose, content_type) VALUES (?, ?, ?, ?, ?, ?)", photo.PhotoID, original_image_name, photoDimensions.Width, photoDimensions.Height, models.PhotoOriginal, contentType)
	if err != nil {
		log.Printf("Could not insert original photo url: %d, %s\n", photo.PhotoID, photoName)
		return err
	}

	return nil
}
