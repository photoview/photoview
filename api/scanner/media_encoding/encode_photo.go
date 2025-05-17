package media_encoding

import (
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"time"

	"github.com/disintegration/imaging"
	"github.com/kkovaletp/photoview/api/graphql/models"
	"github.com/kkovaletp/photoview/api/scanner/media_encoding/executable_worker"
	"github.com/kkovaletp/photoview/api/scanner/media_encoding/media_utils"
	"github.com/kkovaletp/photoview/api/scanner/media_type"
	"github.com/pkg/errors"
	"gopkg.in/vansante/go-ffprobe.v2"

	_ "github.com/strukturag/libheif/go/heif"

	"gorm.io/gorm"
)

func EncodeThumbnail(db *gorm.DB, inputPath string, outputPath string) (*media_utils.PhotoDimensions, error) {

	var siteInfo models.SiteInfo
	if err := db.First(&siteInfo).Error; err != nil {
		return nil, err
	}

	inputImage, err := imaging.Open(inputPath, imaging.AutoOrientation(true))
	if err != nil {
		return nil, err
	}

	dimensions := media_utils.PhotoDimensionsFromRect(inputImage.Bounds())
	dimensions = dimensions.ThumbnailScale()

	thumbImage := imaging.Resize(inputImage, dimensions.Width, dimensions.Height, imaging.NearestNeighbor)
	if err = encodeImageJPEG(thumbImage, outputPath, 60); err != nil {
		return nil, err
	}

	return &dimensions, nil
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
