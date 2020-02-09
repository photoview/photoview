package scanner

import (
	"database/sql"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/nfnt/resize"
	"github.com/viktorstrate/photoview/api/graphql/models"

	// Image decoders
	_ "golang.org/x/image/bmp"
	// _ "golang.org/x/image/tiff"
	_ "github.com/nf/cr2"
	_ "golang.org/x/image/webp"
	_ "image/gif"
	_ "image/png"
)

func ProcessImage(tx *sql.Tx, photoPath string, albumId int) error {

	log.Printf("Processing image: %s\n", photoPath)

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

	thumbFile, err := os.Open(photoPath)
	if err != nil {
		return err
	}
	defer thumbFile.Close()

	image, _, err := image.Decode(thumbFile)
	if err != nil {
		log.Println("ERROR: decoding image")
		return err
	}

	_, err = tx.Exec("INSERT INTO photo_url (photo_id, photo_name, width, height, purpose) VALUES (?, ?, ?, ?, ?)", photo_id, photoName, image.Bounds().Max.X, image.Bounds().Max.Y, models.PhotoOriginal)
	if err != nil {
		log.Printf("Could not insert original photo url: %d, %s\n", photo_id, photoName)
		return err
	}

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
	// Generate image token name
	thumbnailToken := generateToken()

	// Save thumbnail as jpg
	thumbnail_name := fmt.Sprintf("thumbnail_%s_%s", photoName, thumbnailToken)
	thumbnail_name = strings.ReplaceAll(thumbnail_name, ".", "_")
	thumbnail_name = strings.ReplaceAll(thumbnail_name, " ", "_")
	thumbnail_name = thumbnail_name + ".jpg"

	thumbFile, err = os.Create(path.Join(albumCachePath, thumbnail_name))
	if err != nil {
		log.Println("ERROR: Could not make thumbnail file")
		return err
	}
	defer thumbFile.Close()

	jpeg.Encode(thumbFile, thumbnailImage, &jpeg.Options{Quality: 70})

	thumbSize := thumbnailImage.Bounds().Max
	_, err = tx.Exec("INSERT INTO photo_url (photo_id, photo_name, width, height, purpose) VALUES (?, ?, ?, ?, ?)", photo_id, thumbnail_name, thumbSize.X, thumbSize.Y, models.PhotoThumbnail)
	if err != nil {
		return err
	}

	return nil
}

func generateToken() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 8

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
