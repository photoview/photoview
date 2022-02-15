package scanner

import (
	"log"
	"os"
	"path"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/scanner_cache"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func ScanMedia(tx *gorm.DB, mediaPath string, albumId int, cache *scanner_cache.AlbumScannerCache) (*models.Media, bool, error) {
	mediaName := path.Base(mediaPath)

	// Check if media already exists
	{
		var media []*models.Media

		result := tx.Where("path_hash = ?", models.MD5Hash(mediaPath)).Find(&media)

		if result.Error != nil {
			return nil, false, errors.Wrap(result.Error, "scan media fetch from database")
		}

		if result.RowsAffected > 0 {
			// log.Printf("Media already scanned: %s\n", mediaPath)
			return media[0], false, nil
		}
	}

	log.Printf("Scanning media: %s\n", mediaPath)

	mediaType, err := cache.GetMediaType(mediaPath)
	if err != nil {
		return nil, false, errors.Wrap(err, "could determine if media was photo or video")
	}

	var mediaTypeText models.MediaType

	if mediaType.IsVideo() {
		mediaTypeText = models.MediaTypeVideo
	} else {
		mediaTypeText = models.MediaTypePhoto
	}

	stat, err := os.Stat(mediaPath)
	if err != nil {
		return nil, false, err
	}

	media := models.Media{
		Title:    mediaName,
		Path:     mediaPath,
		AlbumID:  albumId,
		Type:     mediaTypeText,
		DateShot: stat.ModTime(),
	}

	if err := tx.Create(&media).Error; err != nil {
		return nil, false, errors.Wrap(err, "could not insert media into database")
	}

	if media.Type == models.MediaTypeVideo {
		if err = ScanVideoMetadata(tx, &media); err != nil {
			log.Printf("WARN: ScanVideoMetadata for %s failed: %s\n", mediaName, err)
		}
	}

	return &media, true, nil
}
