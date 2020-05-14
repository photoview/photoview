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

func ProcessPhoto(tx *sql.Tx, photo *models.Photo) error {

	log.Printf("Processing photo: %s\n", photo.Path)

	imageData := EncodeImageData{
		photo: photo,
	}

	photoName := path.Base(photo.Path)

	photoBaseName := photoName[0 : len(photoName)-len(path.Ext(photoName))]
	photoBaseExt := path.Ext(photoName)

	photoUrlFromDB, err := makePhotoURLChecker(tx, photo.PhotoID)
	if err != nil {
		return err
	}

	// original photo url
	origURL, err := photoUrlFromDB(models.PhotoOriginal)
	if err != nil {
		return err
	}

	if origURL == nil {
		original_image_name := fmt.Sprintf("%s_%s", photoBaseName, utils.GenerateToken())
		original_image_name = strings.ReplaceAll(original_image_name, " ", "_") + photoBaseExt

		photoImage, err := imageData.PhotoImage(tx)
		if err != nil {
			return err
		}

		contentType, err := imageData.ContentType()
		if err != nil {
			return err
		}

		photoDimensions := photoImage.Bounds().Max

		_, err = tx.Exec("INSERT INTO photo_url (photo_id, photo_name, width, height, purpose, content_type) VALUES (?, ?, ?, ?, ?, ?)", photo.PhotoID, original_image_name, photoDimensions.X, photoDimensions.Y, models.PhotoOriginal, contentType)
		if err != nil {
			log.Printf("Could not insert original photo url: %d, %s\n", photo.PhotoID, photoName)
			return err
		}
	}

	// Thumbnail
	thumbURL, err := photoUrlFromDB(models.PhotoThumbnail)
	if err != nil {
		return errors.Wrap(err, "error processing thumbnail")
	}

	// Highres
	highResURL, err := photoUrlFromDB(models.PhotoHighRes)
	if err != nil {
		return errors.Wrap(err, "error processing highres")
	}

	// Make sure photo cache directory exists
	photoCachePath, err := makePhotoCacheDir(photo)
	if err != nil {
		return errors.Wrap(err, "cache directory error")
	}

	// Save thumbnail to cache
	if thumbURL == nil {
		thumbnail_name := fmt.Sprintf("thumbnail_%s_%s", photoName, utils.GenerateToken())
		thumbnail_name = strings.ReplaceAll(thumbnail_name, ".", "_")
		thumbnail_name = strings.ReplaceAll(thumbnail_name, " ", "_")
		thumbnail_name = thumbnail_name + ".jpg"

		thumbnailImage, err := imageData.ThumbnailImage(tx)
		if err != nil {
			return err
		}

		err = imageData.EncodeImageJPEG(tx, path.Join(*photoCachePath, thumbnail_name), 70)
		if err != nil {
			return errors.Wrap(err, "could not create thumbnail cached image")
		}

		thumbSize := thumbnailImage.Bounds().Max
		_, err = tx.Exec("INSERT INTO photo_url (photo_id, photo_name, width, height, purpose, content_type) VALUES (?, ?, ?, ?, ?, ?)", photo.PhotoID, thumbnail_name, thumbSize.X, thumbSize.Y, models.PhotoThumbnail, "image/jpeg")
		if err != nil {
			return err
		}
	} else {
		// Verify that thumbnail photo still exists in cache
		thumbPath := path.Join(*photoCachePath, thumbURL.PhotoName)

		if _, err := os.Stat(thumbPath); os.IsNotExist(err) {
			fmt.Printf("Thumbnail photo found in database but not in cache, re-encoding photo to cache: %s\n", thumbURL.PhotoName)

			err = imageData.EncodeImageJPEG(tx, thumbPath, 70)
			if err != nil {
				log.Println("ERROR: creating thumbnail cached image")
				return err
			}
		}
	}

	// Generate high res jpeg
	if highResURL == nil {

		contentType, err := imageData.ContentType()
		if err != nil {
			return err
		}

		original_web_safe := false
		for _, web_mime := range WebMimetypes {
			if *contentType == web_mime {
				original_web_safe = true
				break
			}
		}

		if !original_web_safe {
			highres_name := fmt.Sprintf("highres_%s_%s", photoName, utils.GenerateToken())
			highres_name = strings.ReplaceAll(highres_name, ".", "_")
			highres_name = strings.ReplaceAll(highres_name, " ", "_")
			highres_name = highres_name + ".jpg"

			photoImage, err := imageData.PhotoImage(tx)
			if err != nil {
				return err
			}

			err = imageData.EncodeImageJPEG(tx, path.Join(*photoCachePath, highres_name), 70)
			if err != nil {
				return errors.Wrap(err, "creating high-res cached image")
			}

			_, err = tx.Exec("INSERT INTO photo_url (photo_id, photo_name, width, height, purpose, content_type) VALUES (?, ?, ?, ?, ?, ?)",
				photo.PhotoID, highres_name, photoImage.Bounds().Max.X, photoImage.Bounds().Max.Y, models.PhotoHighRes, "image/jpeg")
			if err != nil {
				log.Printf("Could not insert highres photo url: %d, %s\n", photo.PhotoID, photoName)
				return err
			}
		}
	} else {
		// Verify that highres photo still exists in cache
		highResPath := path.Join(*photoCachePath, highResURL.PhotoName)

		if _, err := os.Stat(highResPath); os.IsNotExist(err) {
			fmt.Printf("High-res photo found in database but not in cache, re-encoding photo to cache: %s\n", highResURL.PhotoName)

			err = imageData.EncodeImageJPEG(tx, highResPath, 70)
			if err != nil {
				return errors.Wrap(err, "could create high-res cached image")
			}
		}
	}

	return nil
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
