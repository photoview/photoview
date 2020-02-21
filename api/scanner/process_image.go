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

func ProcessImage(tx *sql.Tx, photoPath string, albumId int, content_type string) error {

	// log.Printf("Processing image: %s\n", photoPath)

	photoName := path.Base(photoPath)

	// Check if image already exists
	row := tx.QueryRow("SELECT (photo_id) FROM photo WHERE path = ?", photoPath)
	var photo_id int64
	if err := row.Scan(&photo_id); err != sql.ErrNoRows {
		if err == nil {
			log.Printf("Image already processed: %s\n", photoPath)
			return nil
		} else {
			return err
		}
	}

	result, err := tx.Exec("INSERT INTO photo (title, path, album_id) VALUES (?, ?, ?)", photoName, photoPath, albumId)
	if err != nil {
		log.Printf("ERROR: Could not insert photo into database")
		return err
	}
	photo_id, err = result.LastInsertId()
	if err != nil {
		return err
	}

	photo_file, err := os.Open(photoPath)
	if err != nil {
		return err
	}
	defer photo_file.Close()

	image, _, err := image.Decode(photo_file)
	if err != nil {
		log.Println("ERROR: decoding image")
		return err
	}

	photoBaseName := photoName[0 : len(photoName)-len(path.Ext(photoName))]
	photoBaseExt := path.Ext(photoName)

	// original photo url
	original_image_name := fmt.Sprintf("%s_%s", photoBaseName, utils.GenerateToken())
	original_image_name = strings.ReplaceAll(original_image_name, " ", "_") + photoBaseExt

	_, err = tx.Exec("INSERT INTO photo_url (photo_id, photo_name, width, height, purpose, content_type) VALUES (?, ?, ?, ?, ?, ?)", photo_id, original_image_name, image.Bounds().Max.X, image.Bounds().Max.Y, models.PhotoOriginal, content_type)
	if err != nil {
		log.Printf("Could not insert original photo url: %d, %s\n", photo_id, photoName)
		return err
	}

	// Thumbnail
	thumbnailImage := resize.Thumbnail(1024, 1024, image, resize.Bilinear)

	if _, err := os.Stat("image-cache"); os.IsNotExist(err) {
		if err := os.Mkdir("image-cache", os.ModePerm); err != nil {
			log.Println("ERROR: Could not make image cache directory")
			return err
		}
	}

	// Make album cache dir
	albumCachePath := path.Join("image-cache", strconv.Itoa(albumId))
	if _, err := os.Stat(albumCachePath); os.IsNotExist(err) {
		if err := os.Mkdir(albumCachePath, os.ModePerm); err != nil {
			log.Println("ERROR: Could not make album image cache directory")
			return err
		}
	}

	// Make photo cache dir
	photoCachePath := path.Join(albumCachePath, strconv.Itoa(int(photo_id)))
	if _, err := os.Stat(photoCachePath); os.IsNotExist(err) {
		if err := os.Mkdir(photoCachePath, os.ModePerm); err != nil {
			log.Println("ERROR: Could not make photo image cache directory")
			return err
		}
	}

	// Save thumbnail as jpg
	thumbnail_name := fmt.Sprintf("thumbnail_%s_%s", photoName, utils.GenerateToken())
	thumbnail_name = strings.ReplaceAll(thumbnail_name, ".", "_")
	thumbnail_name = strings.ReplaceAll(thumbnail_name, " ", "_")
	thumbnail_name = thumbnail_name + ".jpg"

	photo_file, err = os.Create(path.Join(photoCachePath, thumbnail_name))
	if err != nil {
		log.Println("ERROR: Could not make thumbnail file")
		return err
	}
	defer photo_file.Close()

	jpeg.Encode(photo_file, thumbnailImage, &jpeg.Options{Quality: 70})

	thumbSize := thumbnailImage.Bounds().Max
	_, err = tx.Exec("INSERT INTO photo_url (photo_id, photo_name, width, height, purpose, content_type) VALUES (?, ?, ?, ?, ?, ?)", photo_id, thumbnail_name, thumbSize.X, thumbSize.Y, models.PhotoThumbnail, "image/jpeg")
	if err != nil {
		return err
	}

	// high res
	original_web_safe := false
	for _, web_mime := range WebMimetypes {
		if content_type == web_mime {
			original_web_safe = true
			break
		}
	}

	// Generate high res jpeg
	if !original_web_safe {
		highres_name := fmt.Sprintf("highres_%s_%s", photoName, utils.GenerateToken())
		highres_name = strings.ReplaceAll(highres_name, ".", "_")
		highres_name = strings.ReplaceAll(highres_name, " ", "_")
		highres_name = highres_name + ".jpg"

		photo_file, err = os.Create(path.Join(photoCachePath, highres_name))
		if err != nil {
			log.Println("ERROR: Could not make highres file")
			return err
		}
		defer photo_file.Close()

		jpeg.Encode(photo_file, image, &jpeg.Options{Quality: 70})

		_, err = tx.Exec("INSERT INTO photo_url (photo_id, photo_name, width, height, purpose, content_type) VALUES (?, ?, ?, ?, ?, ?)", photo_id, highres_name, image.Bounds().Max.X, image.Bounds().Max.Y, models.PhotoHighRes, "image/jpeg")
		if err != nil {
			log.Printf("Could not insert highres photo url: %d, %s\n", photo_id, photoName)
			return err
		}
	}

	return nil
}
