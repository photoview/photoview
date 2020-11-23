package scanner

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"io"
	"log"
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/viktorstrate/photoview/api/graphql/models"
	"gorm.io/gorm"
)

func scanForSideCarFile(path string) *string {
	testPath := path + ".xmp"
	_, err := os.Stat(testPath)

	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		// unexpected error logging
		log.Printf("ERROR: %s", err)
		return nil
	}
	return &testPath

}

func hashSideCarFile(path *string) *string {
	if path == nil {
		return nil
	}

	f, err := os.Open(*path)
	if err != nil {
		log.Printf("ERROR: %s", err)
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Printf("ERROR: %s", err)
	}
	hash := hex.EncodeToString(h.Sum(nil))
	return &hash
}

func ScanMedia(tx *gorm.DB, mediaPath string, albumId uint, cache *AlbumScannerCache) (*models.Media, bool, error) {
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

	var sideCarPath *string
	sideCarPath = nil
	var sideCarHash *string
	sideCarHash = nil
	if mediaType.isVideo() {
		mediaTypeText = "video"
	} else {
		mediaTypeText = "photo"
		// search for sidecar files
		if mediaType.isRaw() {
			sideCarPath = scanForSideCarFile(mediaPath)
			if sideCarPath != nil {
				sideCarHash = hashSideCarFile(sideCarPath)
			}
		}
	}

	stat, err := os.Stat(mediaPath)
	if err != nil {
		return nil, false, err
	}

	result, err := tx.Exec("INSERT INTO media (title, path, path_hash, side_car_path, side_car_hash, album_id, media_type, date_shot) VALUES (?, ?, MD5(path), ?, ?, ?, ?, ?)", mediaName, mediaPath, sideCarPath, sideCarHash, albumId, mediaTypeText, stat.ModTime())
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
