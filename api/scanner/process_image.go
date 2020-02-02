package scanner

import (
	"database/sql"
	"image"
	"image/jpeg"
	"log"
	"math/rand"
	"os"
	"path"
	"strconv"

	"github.com/nfnt/resize"

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
	var id int
	if err := row.Scan(&id); err != sql.ErrNoRows {
		if err == nil {
			log.Printf("Image already processed: %s\n", photoPath)
			return nil
		} else {
			return err
		}
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
	originalToken := generateToken()

	// Save thumbnail as jpg
	thumbFile, err = os.Create(path.Join(albumCachePath, thumbnailToken+".jpg"))
	if err != nil {
		log.Println("ERROR: Could not make thumbnail file")
		return err
	}
	defer thumbFile.Close()

	jpeg.Encode(thumbFile, thumbnailImage, &jpeg.Options{Quality: 70})

	thumbSize := thumbnailImage.Bounds().Max
	thumbRes, err := tx.Exec("INSERT INTO photo_url (token, width, height) VALUES (?, ?, ?)", thumbnailToken, thumbSize.X, thumbSize.Y)
	if err != nil {
		return err
	}
	thumbUrlId, err := thumbRes.LastInsertId()
	if err != nil {
		return err
	}

	origSize := image.Bounds().Max
	origRes, err := tx.Exec("INSERT INTO photo_url (token, width, height) VALUES (?, ?, ?)", originalToken, origSize.X, origSize.Y)
	if err != nil {
		return err
	}
	origUrlId, err := origRes.LastInsertId()

	_, err = tx.Exec("INSERT INTO photo (title, path, album_id, original_url, thumbnail_url) VALUES (?, ?, ?, ?, ?)", photoName, photoPath, albumId, origUrlId, thumbUrlId)
	if err != nil {
		log.Printf("ERROR: Could not insert photo into database")
		return err
	}

	return nil
}

func generateToken() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 24

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
