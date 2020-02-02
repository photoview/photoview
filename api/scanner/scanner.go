package scanner

import (
	"container/list"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/h2non/filetype"
	"github.com/viktorstrate/photoview/api/graphql/models"
)

func ScanUser(database *sql.DB, userId string) error {

	row := database.QueryRow("SELECT * FROM user WHERE user_id = ?", userId)
	user, err := models.NewUserFromRow(row)
	if err != nil {
		log.Printf("Could not find user to scan: %s\n", err.Error())
		return err
	}

	log.Printf("Starting scan for user '%s'\n", user.Username)
	go scan(database, user)

	return nil
}

func scan(database *sql.DB, user *models.User) {
	type scanInfo struct {
		path     string
		parentId *int
	}

	scanQueue := list.New()
	scanQueue.PushBack(scanInfo{
		path:     user.RootPath,
		parentId: nil,
	})

	for scanQueue.Front() != nil {
		albumInfo := scanQueue.Front().Value.(scanInfo)
		scanQueue.Remove(scanQueue.Front())

		albumPath := albumInfo.path
		albumParentId := albumInfo.parentId

		// Read path
		dirContent, err := ioutil.ReadDir(albumPath)
		if err != nil {
			log.Printf("Could not read directory: %s\n", err.Error())
			return
		}

		tx, err := database.Begin()
		if err != nil {
			log.Printf("ERROR: Could not begin database transaction: %s\n", err)
			return
		}

		log.Printf("Scanning directory: %s", albumPath)

		// Make album if not exists
		albumTitle := path.Base(albumPath)
		_, err = tx.Exec("INSERT IGNORE INTO album (title, parent_album, owner_id, path) VALUES (?, ?, ?, ?)", albumTitle, albumParentId, user.UserID, albumPath)
		if err != nil {
			fmt.Printf("ERROR: Could not insert album into database: %s\n", err)
			tx.Rollback()
			return
		}

		row := tx.QueryRow("SELECT album_id FROM album WHERE path = ?", albumPath)
		var albumId int
		if err := row.Scan(&albumId); err != nil {
			fmt.Printf("ERROR: Could not get id of album: %s\n", err)
			tx.Rollback()
			return
		}

		// Commit album transaction
		if err := tx.Commit(); err != nil {
			log.Printf("ERROR: Could not commit database transaction: %s\n", err)
			return
		}

		// Scan for photos
		for _, item := range dirContent {
			photoPath := path.Join(albumPath, item.Name())

			if !item.IsDir() && isPathImage(photoPath) {
				tx, err := database.Begin()
				if err != nil {
					log.Printf("ERROR: Could not begin database transaction for image %s: %s\n", photoPath, err)
					return
				}

				if err := ProcessImage(tx, photoPath, albumId); err != nil {
					log.Printf("ERROR: processing image %s: %s", photoPath, err)
					tx.Rollback()
					return
				}

				tx.Commit()
			}
		}

		// Scan for sub-albums
		for _, item := range dirContent {
			subalbumPath := path.Join(albumPath, item.Name())

			if item.IsDir() && directoryContainsPhotos(subalbumPath) {
				scanQueue.PushBack(scanInfo{
					path:     subalbumPath,
					parentId: &albumId,
				})
			}
		}
	}

	log.Println("Done scanning")
}

func directoryContainsPhotos(rootPath string) bool {
	scanQueue := list.New()
	scanQueue.PushBack(rootPath)

	for scanQueue.Front() != nil {

		dirPath := scanQueue.Front().Value.(string)
		scanQueue.Remove(scanQueue.Front())

		dirContent, err := ioutil.ReadDir(dirPath)
		if err != nil {
			log.Printf("Could not read directory: %s\n", err.Error())
			return false
		}

		for _, fileInfo := range dirContent {
			filePath := path.Join(dirPath, fileInfo.Name())
			if fileInfo.IsDir() {
				scanQueue.PushBack(filePath)
			} else {
				if isPathImage(filePath) {
					return true
				}
			}
		}

	}

	return false
}

var supported_mimetypes = [...]string{
	"image/jpeg",
	"image/png",
	"image/tiff",
	"image/webp",
	"image/x-canon-cr2",
	"image/bmp",
}

func isPathImage(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		log.Printf("Could not open file %s: %s\n", path, err)
		return false
	}
	defer file.Close()

	head := make([]byte, 261)
	if _, err := file.Read(head); err != nil {
		log.Printf("Could not read file %s: %s\n", path, err)
		return false
	}

	imgType, err := filetype.Image(head)
	if err != nil {
		return false
	}

	for _, supported_mime := range supported_mimetypes {
		if supported_mime == imgType.MIME.Value {
			return true
		}
	}

	log.Printf("Unsupported image %s of type %s\n", path, imgType.MIME.Value)
	return false
}
