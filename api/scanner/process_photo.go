package scanner

import (
	"database/sql"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/nfnt/resize"
	"github.com/viktorstrate/photoview/api/graphql/models"
	"github.com/viktorstrate/photoview/api/utils"

	// Image decoders
	_ "golang.org/x/image/bmp"
	// _ "golang.org/x/image/tiff"
	_ "image/gif"
	_ "image/png"

	_ "github.com/nf/cr2"
	_ "golang.org/x/image/webp"
)

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

func ProcessPhoto(tx *sql.Tx, photo *models.Photo, content_type *string) error {

	log.Printf("Processing photo: %s\n", photo.Path)

	imageData := processImageData{
		photoPath: photo.Path,
	}

	photoName := path.Base(photo.Path)

	photoBaseName := photoName[0 : len(photoName)-len(path.Ext(photoName))]
	photoBaseExt := path.Ext(photoName)

	photoChecker, err := makePhotoURLChecker(tx, photo.PhotoID)
	if err != nil {
		return err
	}

	// original photo url
	origURL, err := photoChecker(models.PhotoOriginal)
	if err != nil {
		return err
	}

	if origURL == nil {
		original_image_name := fmt.Sprintf("%s_%s", photoBaseName, utils.GenerateToken())
		original_image_name = strings.ReplaceAll(original_image_name, " ", "_") + photoBaseExt

		photoImage, err := imageData.PhotoImage()
		if err != nil {
			return err
		}

		_, err = tx.Exec("INSERT INTO photo_url (photo_id, photo_name, width, height, purpose, content_type) VALUES (?, ?, ?, ?, ?, ?)", photo.PhotoID, original_image_name, photoImage.Bounds().Max.X, photoImage.Bounds().Max.Y, models.PhotoOriginal, content_type)
		if err != nil {
			log.Printf("Could not insert original photo url: %d, %s\n", photo.PhotoID, photoName)
			return err
		}
	}

	// Thumbnail
	thumbURL, err := photoChecker(models.PhotoThumbnail)
	if err != nil {
		return err
	}

	// Highres
	highResURL, err := photoChecker(models.PhotoHighRes)
	if err != nil {
		return err
	}

	// Make sure photo cache directory exists
	photoCachePath, err := makePhotoCacheDir(photo)
	if err != nil {
		return err
	}

	// Save thumbnail to cache
	if thumbURL == nil {
		thumbnail_name := fmt.Sprintf("thumbnail_%s_%s", photoName, utils.GenerateToken())
		thumbnail_name = strings.ReplaceAll(thumbnail_name, ".", "_")
		thumbnail_name = strings.ReplaceAll(thumbnail_name, " ", "_")
		thumbnail_name = thumbnail_name + ".jpg"

		thumbnailImage, err := imageData.ThumbnailImage()
		if err != nil {
			return err
		}

		err = encodeImageJPEG(path.Join(*photoCachePath, thumbnail_name), thumbnailImage, &jpeg.Options{Quality: 70})
		if err != nil {
			log.Println("ERROR: creating high-res cached image")
			return err
		}

		thumbSize := thumbnailImage.Bounds().Max
		_, err = tx.Exec("INSERT INTO photo_url (photo_id, photo_name, width, height, purpose, content_type) VALUES (?, ?, ?, ?, ?, ?)", photo.PhotoID, thumbnail_name, thumbSize.X, thumbSize.Y, models.PhotoThumbnail, "image/jpeg")
		if err != nil {
			return err
		}
	} else if thumbURL != nil {
		thumbPath := path.Join(*photoCachePath, thumbURL.PhotoName)

		if _, err := os.Stat(thumbPath); os.IsNotExist(err) {
			fmt.Printf("Thumbnail photo found in database but not in cache, re-encoding photo to cache: %s\n", thumbURL.PhotoName)

			thumbnailImage, err := imageData.ThumbnailImage()
			if err != nil {
				return err
			}

			err = encodeImageJPEG(thumbPath, thumbnailImage, &jpeg.Options{Quality: 70})
			if err != nil {
				log.Println("ERROR: creating thumbnail cached image")
				return err
			}
		}
	}

	// high res
	original_web_safe := false
	for _, web_mime := range WebMimetypes {
		if *content_type == web_mime {
			original_web_safe = true
			break
		}
	}

	// Generate high res jpeg
	if highResURL == nil {
		if !original_web_safe {
			highres_name := fmt.Sprintf("highres_%s_%s", photoName, utils.GenerateToken())
			highres_name = strings.ReplaceAll(highres_name, ".", "_")
			highres_name = strings.ReplaceAll(highres_name, " ", "_")
			highres_name = highres_name + ".jpg"

			photoImage, err := imageData.PhotoImage()
			if err != nil {
				return err
			}

			err = encodeImageJPEG(path.Join(*photoCachePath, highres_name), photoImage, &jpeg.Options{Quality: 70})
			if err != nil {
				log.Println("ERROR: creating high-res cached image")
				return err
			}

			_, err = tx.Exec("INSERT INTO photo_url (photo_id, photo_name, width, height, purpose, content_type) VALUES (?, ?, ?, ?, ?, ?)",
				photo.PhotoID, highres_name, photoImage.Bounds().Max.X, photoImage.Bounds().Max.Y, models.PhotoHighRes, "image/jpeg")
			if err != nil {
				log.Printf("Could not insert highres photo url: %d, %s\n", photo.PhotoID, photoName)
				return err
			}
		}
	} else if highResURL != nil {
		highResPath := path.Join(*photoCachePath, highResURL.PhotoName)

		if _, err := os.Stat(highResPath); os.IsNotExist(err) {
			fmt.Printf("High-res photo found in database but not in cache, re-encoding photo to cache: %s\n", highResURL.PhotoName)

			photoImage, err := imageData.PhotoImage()
			if err != nil {
				return err
			}

			err = encodeImageJPEG(highResPath, photoImage, &jpeg.Options{Quality: 70})
			if err != nil {
				log.Println("ERROR: creating high-res cached image")
				return err
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

func encodeImageJPEG(photoPath string, photoImage image.Image, jpegOptions *jpeg.Options) error {
	photo_file, err := os.Create(photoPath)
	if err != nil {
		log.Printf("ERROR: Could not create file: %s\n", photoPath)
		return err
	}
	defer photo_file.Close()

	err = jpeg.Encode(photo_file, photoImage, jpegOptions)
	if err != nil {
		return err
	}

	return nil
}

type processImageData struct {
	photoPath       string
	_photoImage     image.Image
	_thumbnailImage image.Image
}

func (img *processImageData) PhotoImage() (image.Image, error) {
	if img._photoImage != nil {
		return img._photoImage, nil
	}

	photoFile, err := os.Open(img.photoPath)
	if err != nil {
		return nil, err
	}
	defer photoFile.Close()

	photoImg, _, err := image.Decode(photoFile)
	if err != nil {
		log.Println("ERROR: decoding image")
		return nil, err
	}

	img._photoImage = photoImg
	return img._photoImage, nil
}

func (img *processImageData) ThumbnailImage() (image.Image, error) {
	photoImage, err := img.PhotoImage()
	if err != nil {
		return nil, err
	}

	thumbImage := resize.Thumbnail(1024, 1024, photoImage, resize.Bilinear)
	img._thumbnailImage = thumbImage

	return img._thumbnailImage, nil
}
