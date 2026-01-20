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
	"github.com/photoview/photoview/api/scanner/scanner_utils"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"gorm.io/gorm"
)

var ProcessSingleMediaFunc = ProcessSingleMedia

func ScanMedia(tx *gorm.DB, fs afero.Fs, mediaPath string, localMediaPath string, albumId int, cache *scanner_cache.AlbumScannerCache) (*models.Media, bool, error) {
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

	mediaType := cache.GetMediaType(localMediaPath)
	if mediaType == media_type.TypeUnknown {
		return nil, false, fmt.Errorf("could not determine if media %s of type %s was photo or video", mediaPath, mediaType)
	}

	var mediaTypeText models.MediaType

	if mediaType.IsVideo() {
		mediaTypeText = models.MediaTypeVideo
	} else {
		mediaTypeText = models.MediaTypePhoto
	}

	// Download to temporary local path if needed
	localMediaPath, err := scanner_utils.DownloadToLocalIfNeeded(fs, mediaPath)
	if err != nil {
		return nil, false, errors.Wrapf(err, "could not download local media path: %s", mediaPath)
	}

	stat, err := fs.Stat(mediaPath)
	if err != nil {
		return nil, false, err
	}

	media := models.Media{
		Title:     mediaName,
		Path:      mediaPath,
		LocalPath: localMediaPath,
		AlbumID:   albumId,
		Type:      mediaTypeText,
		DateShot:  stat.ModTime(),
	}

	if err := tx.Create(&media).Error; err != nil {
		return nil, false, errors.Wrap(err, "could not insert media into database")
	}

	return &media, true, nil
}

// ProcessSingleMedia processes a single media, might be used to reprocess media with corrupted cache
// Function waits for processing to finish before returning.
func ProcessSingleMedia(ctx context.Context, db *gorm.DB, fs afero.Fs, cacheFs afero.Fs, media *models.Media) error {
	albumCache := scanner_cache.MakeAlbumCache(fs)

	var album models.Album
	if err := db.Model(media).Association("Album").Find(&album); err != nil {
		return err
	}

	// FIXME: Download to local path ?

	mediaData := media_encoding.NewEncodeMediaData(media)

	taskContext := scanner_task.NewTaskContext(ctx, db, fs, cacheFs, &album, albumCache)
	if err := scanMedia(taskContext, media, &mediaData, 0, 1); err != nil {
		return errors.Wrap(err, "single media scan")
	}

	return nil
}
