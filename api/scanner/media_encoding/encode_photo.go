package media_encoding

import (
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/media_encoding/executable_worker"
	"github.com/photoview/photoview/api/scanner/media_encoding/media_utils"
	"github.com/photoview/photoview/api/scanner/media_type"
	"github.com/pkg/errors"
	"gopkg.in/vansante/go-ffprobe.v2"

	_ "github.com/strukturag/libheif/go/heif"

	"gorm.io/gorm"
)

func EncodeThumbnail(db *gorm.DB, inputPath string, outputPath string) (*media_utils.PhotoDimensions, error) {
	if err := executable_worker.Magick.GenerateThumbnail(inputPath, outputPath, 1024, 1024); err != nil {
		return nil, fmt.Errorf("can't generate thumbnail of flie %q: %w", inputPath, err)
	}

	ret, err := executable_worker.Magick.IdentifyDimension(outputPath)
	if err != nil {
		return nil, fmt.Errorf("can't identify dimension of file %q: %w", outputPath, err)
	}

	return &media_utils.PhotoDimensions{
		Width:  ret.Width,
		Height: ret.Height,
	}, nil
}

func encodeImageJPEG(image image.Image, outputPath string, jpegQuality int) error {
	photo_file, err := os.Create(outputPath)
	if err != nil {
		return errors.Wrapf(err, "could not create file: %s", outputPath)
	}
	defer photo_file.Close()

	err = jpeg.Encode(photo_file, image, &jpeg.Options{Quality: jpegQuality})
	if err != nil {
		return err
	}

	return nil
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
