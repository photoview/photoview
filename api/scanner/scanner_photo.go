package scanner

import (
	"database/sql"
	"log"
	"path"

	"github.com/viktorstrate/photoview/api/graphql/models"
)

func ScanPhoto(tx *sql.Tx, photoPath string, albumId int) (*models.Photo, bool, error) {
	photoName := path.Base(photoPath)

	// Check if image already exists
	{
		row := tx.QueryRow("SELECT * FROM photo WHERE path_hash = MD5(?)", photoPath)
		photo, err := models.NewPhotoFromRow(row)
		if err != sql.ErrNoRows {
			if err == nil {
				log.Printf("Image already scanned: %s\n", photoPath)
				return photo, false, nil
			} else {
				return nil, false, err
			}
		}
	}

	log.Printf("Scanning image: %s\n", photoPath)

	result, err := tx.Exec("INSERT INTO photo (title, path, path_hash, album_id) VALUES (?, ?, MD5(path), ?)", photoName, photoPath, albumId)
	if err != nil {
		log.Printf("ERROR: Could not insert photo into database")
		return nil, false, err
	}
	photo_id, err := result.LastInsertId()
	if err != nil {
		return nil, false, err
	}

	row := tx.QueryRow("SELECT * FROM photo WHERE photo_id = ?", photo_id)
	photo, err := models.NewPhotoFromRow(row)
	if err != nil {
		return nil, false, err
	}

	_, err = ScanEXIF(tx, photo)
	if err != nil {
		log.Printf("ERROR: ScanEXIF for %s: %s\n", photoName, err)
	}

	return photo, true, nil
}
