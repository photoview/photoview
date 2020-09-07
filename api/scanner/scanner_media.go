package scanner

import (
	"database/sql"
	"log"
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/viktorstrate/photoview/api/graphql/models"
)

func ScanMedia(tx *sql.Tx, mediaPath string, albumId int, cache *AlbumScannerCache) (*models.Media, bool, error) {
	mediaName := path.Base(mediaPath)

	// Check if image already exists
	{
		row := tx.QueryRow("SELECT * FROM media WHERE path_hash = MD5(?)", mediaPath)
		photo, err := models.NewMediaFromRow(row)
		if err != sql.ErrNoRows {
			if err == nil {
				log.Printf("Media already scanned: %s\n", mediaPath)
				return photo, false, nil
			} else {
				return nil, false, errors.Wrap(err, "scan media fetch from database")
			}
		}
	}

	log.Printf("Scanning media: %s\n", mediaPath)

	mediaType, err := cache.GetMediaType(mediaPath)
	if err != nil {
		return nil, false, errors.Wrap(err, "could determine if media was photo or video")
	}

	var mediaTypeText string
	if mediaType.isVideo() {
		mediaTypeText = "video"
	} else {
		mediaTypeText = "photo"
	}

	stat, err := os.Stat(mediaPath)
	if err != nil {
		return nil, false, err
	}

	result, err := tx.Exec("INSERT INTO media (title, path, path_hash, album_id, media_type, date_shot) VALUES (?, ?, MD5(path), ?, ?, ?)", mediaName, mediaPath, albumId, mediaTypeText, stat.ModTime())
	if err != nil {
		return nil, false, errors.Wrap(err, "could not insert media into database")
	}
	media_id, err := result.LastInsertId()
	if err != nil {
		return nil, false, err
	}

	row := tx.QueryRow("SELECT * FROM media WHERE media_id = ?", media_id)
	media, err := models.NewMediaFromRow(row)
	if err != nil {
		return nil, false, errors.Wrap(err, "failed to get media by id from database")
	}

	_, err = ScanEXIF(tx, media)
	if err != nil {
		log.Printf("WARN: ScanEXIF for %s failed: %s\n", mediaName, err)
	}

	if media.Type == models.MediaTypeVideo {
		if err = ScanVideoMetadata(tx, media); err != nil {
			log.Printf("WARN: ScanVideoMetadata for %s failed: %s\n", mediaName, err)
		}
	}

	return media, true, nil
}
