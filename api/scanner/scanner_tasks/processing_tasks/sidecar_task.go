package processing_tasks

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"path"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/media_encoding"
	"github.com/photoview/photoview/api/scanner/media_type"
	"github.com/photoview/photoview/api/scanner/scanner_task"
	"github.com/photoview/photoview/api/scanner/scanner_utils"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

type SidecarTask struct {
	scanner_task.ScannerTaskBase
}

func (t SidecarTask) AfterMediaFound(ctx scanner_task.TaskContext, media *models.Media, newMedia bool) error {
	if media.Type != models.MediaTypePhoto || !newMedia {
		return nil
	}

	mediaType := ctx.GetCache().GetMediaType(media.Path)
	if mediaType == media_type.TypeUnknown {
		return fmt.Errorf("scan for sidecar file %s failed: media type is %s", media.Path, mediaType)
	}

	if mediaType.IsWebCompatible() {
		return nil
	}

	var sideCarPath *string = nil
	var sideCarHash *string = nil

	sideCarPath = scanForSideCarFile(ctx.GetFS(), media.Path)
	if sideCarPath != nil {
		sideCarHash = hashSideCarFile(ctx.GetFS(), sideCarPath)
	}

	// Add sidecar data to media
	media.SideCarPath = sideCarPath
	media.SideCarHash = sideCarHash
	if err := ctx.GetDB().Save(media).Error; err != nil {
		if sideCarPath != nil {
			return errors.Wrapf(err, "update media sidecar info (%s)", *sideCarPath)
		}

		return errors.Wrapf(err, "update media sidecar info")
	}

	return nil
}

func (t SidecarTask) ProcessMedia(ctx scanner_task.TaskContext, mediaData *media_encoding.EncodeMediaData, mediaCachePath string) (updatedURLs []*models.MediaURL, err error) {
	fs := ctx.GetFS()

	mediaType, err := mediaData.ContentType(ctx.GetFS())
	if err != nil {
		return []*models.MediaURL{}, errors.Wrap(err, "sidecar task, process media")
	}

	if mediaType.IsWebCompatible() {
		return []*models.MediaURL{}, nil
	}

	photo := mediaData.Media

	sideCarFileHasChanged := false
	var currentFileHash *string
	currentSideCarPath := scanForSideCarFile(ctx.GetFS(), photo.Path)

	if currentSideCarPath != nil {
		currentFileHash = hashSideCarFile(ctx.GetFS(), currentSideCarPath)
		if currentFileHash == nil {
			return []*models.MediaURL{}, errors.New("sidecar task, hash sidecar file failed")
		}

		if photo.SideCarHash == nil || *photo.SideCarHash != *currentFileHash {
			sideCarFileHasChanged = true
		}
	} else if photo.SideCarPath != nil { // sidecar has been deleted since last scan
		sideCarFileHasChanged = true
	}

	if !sideCarFileHasChanged {
		return []*models.MediaURL{}, nil
	}

	fmt.Printf("Detected changed sidecar file for %s recreating JPG's to reflect changes\n", photo.Path)

	highResURL, err := photo.GetHighRes()
	if err != nil {
		return []*models.MediaURL{}, errors.Wrap(err, "sidecar task, get high-res media_url")
	}

	thumbURL, err := photo.GetThumbnail()
	if err != nil {
		return []*models.MediaURL{}, errors.Wrap(err, "sidecar task, get high-res media_url")
	}

	// update high res image may be cropped so dimentions and file size can change
	baseImagePath := path.Join(mediaCachePath, highResURL.MediaName) // update base image path for thumbnail
	tempHighResPath := baseImagePath + ".hold"
	if err := fs.Rename(baseImagePath, tempHighResPath); err != nil {
		return []*models.MediaURL{}, errors.Wrapf(err, "sidecar task, hold high-res image: %s", baseImagePath)
	}

	updatedHighRes, err := generateSaveHighResJPEG(ctx.GetDB(), ctx.GetFS(), photo, mediaData, highResURL.MediaName, baseImagePath, highResURL)
	if err != nil {
		if restoreErr := fs.Rename(tempHighResPath, baseImagePath); restoreErr != nil {
			log.Printf("ERROR: restoring high-res image failed: %s", restoreErr)
		}
		return []*models.MediaURL{}, errors.Wrap(err, "sidecar task, recreating high-res cached image")
	}
	if err := fs.Remove(tempHighResPath); err != nil {
		log.Printf("ERROR: removing temp high-res image failed: %s", err)
	}

	// update thumbnail image may be cropped so dimentions and file size can change
	thumbPath := path.Join(mediaCachePath, thumbURL.MediaName)
	tempThumbPath := thumbPath + ".hold" // hold onto the original image incase for some reason we fail to recreate one with the new settings
	if err := fs.Rename(thumbPath, tempThumbPath); err != nil {
		return []*models.MediaURL{}, errors.Wrapf(err, "sidecar task, hold thumbnail image: %s", thumbPath)
	}
	updatedThumbnail, err := generateSaveThumbnailJPEG(ctx.GetDB(), ctx.GetFS(), photo, thumbURL.MediaName, mediaCachePath, baseImagePath, thumbURL)
	if err != nil {
		if restoreErr := fs.Rename(tempThumbPath, thumbPath); restoreErr != nil {
			log.Printf("ERROR: restoring thumbnail image failed: %s", restoreErr)
		}
		return []*models.MediaURL{}, errors.Wrap(err, "recreating thumbnail cached image")
	}
	if err := fs.Remove(tempThumbPath); err != nil {
		log.Printf("ERROR: removing temp high-res image failed: %s", err)
	}

	photo.SideCarHash = currentFileHash
	photo.SideCarPath = currentSideCarPath

	// save new side car hash
	if err := ctx.GetDB().Save(&photo).Error; err != nil {
		return []*models.MediaURL{}, errors.Wrapf(err, "could not update side car hash for media: %s", photo.Path)
	}

	return []*models.MediaURL{
		updatedThumbnail,
		updatedHighRes,
	}, nil
}

func scanForSideCarFile(fs afero.Fs, path string) *string {
	testPath := path + ".xmp"

	if scanner_utils.FileExists(fs, testPath) {
		return &testPath
	}

	return nil
}

func hashSideCarFile(fs afero.Fs, path *string) *string {
	if path == nil {
		return nil
	}

	f, err := fs.Open(*path)
	if err != nil {
		log.Printf("ERROR: %s", err)
		return nil
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Printf("ERROR: %s", err)
		return nil
	}
	hash := hex.EncodeToString(h.Sum(nil))
	return &hash
}
