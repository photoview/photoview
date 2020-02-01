package scanner

import (
	"database/sql"
	"path"
)

func ProcessImage(tx *sql.Tx, photoPath string, albumId int) error {
	photoName := path.Base(photoPath)

	_, err := tx.Exec("INSERT IGNORE INTO photo (title, path, album_id) VALUES (?, ?, ?)", photoName, photoPath, albumId)
	if err != nil {
		return err
	}
}
