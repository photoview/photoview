package scanner

import (
	"crypto/md5"
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

	// Check if media already exists
	{
		var media models.Media
		if err := tx.Where("path_hash = MD5(?)", mediaPath).First(&media).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			if err == nil {
				log.Printf("Media already scanned: %s\n", mediaPath)
				return &media, false, nil
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

	var mediaTypeText models.MediaType

	var sideCarPath *string = nil
	var sideCarHash *string = nil

	if mediaType.isVideo() {
		mediaTypeText = models.MediaTypeVideo
	} else {
		mediaTypeText = models.MediaTypePhoto
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

	media := models.Media{
		Title:       mediaName,
		Path:        mediaPath,
		SideCarPath: sideCarPath,
		SideCarHash: sideCarHash,
		AlbumId:     albumId,
		Type:        mediaTypeText,
		DateShot:    stat.ModTime(),
	}

	if err := tx.Create(&media).Error; err != nil {
		return nil, false, errors.Wrap(err, "could not insert media into database")
	}

	_, err = ScanEXIF(tx, &media)
	if err != nil {
		log.Printf("WARN: ScanEXIF for %s failed: %s\n", mediaName, err)
	}

	if media.Type == models.MediaTypeVideo {
		if err = ScanVideoMetadata(tx, &media); err != nil {
			log.Printf("WARN: ScanVideoMetadata for %s failed: %s\n", mediaName, err)
		}
	}

	return &media, true, nil
}
