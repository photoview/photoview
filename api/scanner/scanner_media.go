package scanner

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/exif"
	"github.com/photoview/photoview/api/scanner/media_type"
	"github.com/photoview/photoview/api/scanner/scanner_cache"
	"github.com/photoview/photoview/api/scanner/scanner_utils"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func scanForSideCarFile(path string) *string {
	testPath := path + ".xmp"

	if scanner_utils.FileExists(testPath) {
		return &testPath
	}

	return nil
}

func scanForRawCounterpartFile(imagePath string) *string {
	ext := filepath.Ext(imagePath)
	fileExtType, found := media_type.GetExtensionMediaType(ext)

	if found {
		if !fileExtType.IsBasicTypeSupported() {
			return nil
		}
	}

	rawPath := media_type.RawCounterpart(imagePath)
	if rawPath != nil {
		return rawPath
	}

	return nil
}

func scanForCompressedCounterpartFile(imagePath string) *string {
	ext := filepath.Ext(imagePath)
	fileExtType, found := media_type.GetExtensionMediaType(ext)

	if found {
		if fileExtType.IsBasicTypeSupported() {
			return nil
		}
	}

	pathWithoutExt := strings.TrimSuffix(imagePath, path.Ext(imagePath))
	for _, ext := range media_type.TypeJpeg.FileExtensions() {
		testPath := pathWithoutExt + ext
		if scanner_utils.FileExists(testPath) {
			return &testPath
		}
	}

	return nil
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
			log.Printf("Media already scanned: %s\n", mediaPath)
			return media[0], false, nil
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

	if mediaType.IsVideo() {
		mediaTypeText = models.MediaTypeVideo
	} else {
		mediaTypeText = models.MediaTypePhoto
		// search for sidecar files
		if mediaType.IsRaw() {
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
		AlbumID:     albumId,
		Type:        mediaTypeText,
		DateShot:    stat.ModTime(),
	}

	if err := tx.Create(&media).Error; err != nil {
		return nil, false, errors.Wrap(err, "could not insert media into database")
	}

	_, err = exif.SaveEXIF(tx, &media)
	if err != nil {
		log.Printf("WARN: SaveEXIF for %s failed: %s\n", mediaName, err)
	}

	if media.Type == models.MediaTypeVideo {
		if err = ScanVideoMetadata(tx, &media); err != nil {
			log.Printf("WARN: ScanVideoMetadata for %s failed: %s\n", mediaName, err)
		}
	}

	return &media, true, nil
}
