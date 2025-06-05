package media_encoding

import (
	"context"
	"fmt"
	"image"
	"time"

	"github.com/kkovaletp/photoview/api/graphql/models"
	"github.com/kkovaletp/photoview/api/scanner/media_encoding/executable_worker"
	"github.com/kkovaletp/photoview/api/scanner/media_type"
	"github.com/pkg/errors"
	"gopkg.in/vansante/go-ffprobe.v2"

	"gorm.io/gorm"
)

// Dimension presents the Dimension of a image.
type Dimension struct {
	Width  int
	Height int
}

// ThumbnailScale generates a new dimension for thumbnails.
func (d *Dimension) ThumbnailScale() Dimension {
	if d.Height == 0 || d.Width == 0 {
		return Dimension{Width: 0, Height: 0}
	}

	aspect := float64(d.Width) / float64(d.Height)

	var width, height int

	if aspect > 1 {
		width = 1024
		height = int(1024 / aspect)
	} else {
		width = int(1024 * aspect)
		height = 1024
	}

	if width > d.Width {
		width = d.Width
		height = d.Height
	}

	return Dimension{
		Width:  width,
		Height: height,
	}
}

// GetPhotoDimensions returns the dimension of the image `imagePath`.
func GetPhotoDimensions(imagePath string) (Dimension, error) {
	w, h, err := executable_worker.Magick.IdentifyDimension(imagePath)
	if err != nil {
		return Dimension{}, fmt.Errorf("identify dimension %q error: %w", imagePath, err)
	}

	return Dimension{
		Width:  int(w),
		Height: int(h),
	}, nil
}

// EncodeThumbnail encodes a thumbnail of `inputPath`, and store it as `outputPath`.
// It returns the dimension of the thumbnail. The thumbnail will be not bigger than 1024x1024.
func EncodeThumbnail(db *gorm.DB, inputPath string, outputPath string) (Dimension, error) {
	w, h, err := executable_worker.Magick.IdentifyDimension(inputPath)
	if err != nil {
		return Dimension{}, fmt.Errorf("can't generate thumbnail of file %q: %w", inputPath, err)
	}

	origin := Dimension{
		Width:  int(w),
		Height: int(h),
	}
	thumbnail := origin.ThumbnailScale()

	if err := executable_worker.Magick.GenerateThumbnail(inputPath, outputPath, uint(thumbnail.Width), uint(thumbnail.Height)); err != nil {
		return Dimension{}, fmt.Errorf("can't generate thumbnail of file %q: %w", inputPath, err)
	}

	return thumbnail, nil
}

// EncodeMediaData is used to easily decode media data, with a cache so expensive operations are not repeated
type EncodeMediaData struct {
	Media           *models.Media
	CounterpartPath *string
	_photoImage     image.Image
	_contentType    media_type.MediaType
	_videoMetadata  *ffprobe.ProbeData
}

func NewEncodeMediaData(media *models.Media) EncodeMediaData {
	fileType := media_type.GetMediaType(media.Path)

	return EncodeMediaData{
		Media:        media,
		_contentType: fileType,
	}
}

// ContentType reads the image to determine its content type
func (img *EncodeMediaData) ContentType() (media_type.MediaType, error) {
	if img._contentType != media_type.TypeUnknown {
		return img._contentType, nil
	}

	imgType := media_type.GetMediaType(img.Media.Path)
	if imgType == media_type.TypeUnknown {
		return imgType, fmt.Errorf("unknown type of %q", img.Media.Path)
	}

	img._contentType = imgType
	return imgType, nil
}

func (img *EncodeMediaData) EncodeHighRes(outputPath string) error {
	contentType, err := img.ContentType()
	if err != nil {
		return err
	}

	if !contentType.IsSupported() {
		return errors.New("could not convert photo as file format is not supported")
	}

	// Use magick if there is no counterpart JPEG file to use instead
	if contentType.IsImage() && !contentType.IsWebCompatible() {
		imgPath := img.Media.Path
		if img.CounterpartPath != nil {
			imgPath = *img.CounterpartPath
		}

		err := executable_worker.Magick.EncodeJpeg(imgPath, outputPath, 70)
		if err != nil {
			return fmt.Errorf("failed to convert RAW photo %q to JPEG: %w", imgPath, err)
		}
	}

	return nil
}

func (enc *EncodeMediaData) VideoMetadata() (*ffprobe.ProbeData, error) {

	if enc._videoMetadata != nil {
		return enc._videoMetadata, nil
	}

	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()
	data, err := ffprobe.ProbeURL(ctx, enc.Media.Path)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read video metadata (%s)", enc.Media.Title)
	}

	enc._videoMetadata = data
	return enc._videoMetadata, nil
}
