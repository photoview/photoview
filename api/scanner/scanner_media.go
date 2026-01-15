package scanner

import (
	"context"
	"fmt"
	"log"
	"path"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/media_encoding"
	"github.com/photoview/photoview/api/scanner/media_type"
	"github.com/photoview/photoview/api/scanner/scanner_cache"
	"github.com/photoview/photoview/api/scanner/scanner_task"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"gorm.io/gorm"
)

var ProcessSingleMediaFunc = ProcessSingleMedia

func ScanMedia(tx *gorm.DB, fs afero.Fs, mediaPath string, albumId int, cache *scanner_cache.AlbumScannerCache) (*models.Media, bool, error) {
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

	mediaType := cache.GetMediaType(mediaPath)
	if mediaType == media_type.TypeUnknown {
		return nil, false, fmt.Errorf("could not determine if media %s of type %s was photo or video", mediaPath, mediaType)
	}

	var mediaTypeText models.MediaType

	if mediaType.IsVideo() {
		mediaTypeText = models.MediaTypeVideo
	} else {
		mediaTypeText = models.MediaTypePhoto
	}

	stat, err := fs.Stat(mediaPath)
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

	return &media, true, nil
}

// ProcessSingleMedia processes a single media, might be used to reprocess media with corrupted cache
// Function waits for processing to finish before returning.
func ProcessSingleMedia(ctx context.Context, db *gorm.DB, fs afero.Fs, media *models.Media) error {
	albumCache := scanner_cache.MakeAlbumCache()

	var album models.Album
	if err := db.Model(media).Association("Album").Find(&album); err != nil {
		return err
	}

	mediaData := media_encoding.NewEncodeMediaData(fs, media)

	taskContext := scanner_task.NewTaskContext(ctx, db, fs, &album, albumCache)
	if err := scanMedia(taskContext, media, &mediaData, 0, 1); err != nil {
		return errors.Wrap(err, "single media scan")
	}

	return nil
}
