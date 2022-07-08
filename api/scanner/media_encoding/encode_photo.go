package media_encoding

import (
	"context"
	"image"
	"image/jpeg"
	"os"
	"time"

	"github.com/disintegration/imaging"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/media_encoding/executable_worker"
	"github.com/photoview/photoview/api/scanner/media_encoding/media_utils"
	"github.com/photoview/photoview/api/scanner/media_type"
	"github.com/photoview/photoview/api/utils"
	"github.com/pkg/errors"
	"gopkg.in/vansante/go-ffprobe.v2"

	_ "github.com/strukturag/libheif/go/heif"
)

func EncodeThumbnail(inputPath string, outputPath string) (*media_utils.PhotoDimensions, error) {

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
	_contentType    *media_type.MediaType
	_videoMetadata  *ffprobe.ProbeData
}

func NewEncodeMediaData(media *models.Media) EncodeMediaData {
	return EncodeMediaData{
		Media: media,
	}
}

// ContentType reads the image to determine its content type
func (img *EncodeMediaData) ContentType() (*media_type.MediaType, error) {
	if img._contentType != nil {
		return img._contentType, nil
	}

	imgType, err := media_type.GetMediaType(img.Media.Path)
	if err != nil {
		return nil, err
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

	// Use darktable if there is no counterpart JPEG file to use instead
	if contentType.IsRaw() && img.CounterpartPath == nil {
		if executable_worker.DarktableCli.IsInstalled() {
			err := executable_worker.DarktableCli.EncodeJpeg(img.Media.Path, outputPath, 70)
			if err != nil {
				return err
			}
		} else {
			return errors.New("could not convert photo as no RAW converter was found")
		}
	} else {
		image, err := img.photoImage()
		if err != nil {
			return err
		}

		encodeImageJPEG(image, outputPath, 70)
	}

	return nil
}

// photoImage reads and decodes the image file and saves it in a cache so the photo in only decoded once
func (img *EncodeMediaData) photoImage() (image.Image, error) {
	if img._photoImage != nil {
		return img._photoImage, nil
	}

	var photoPath string
	if img.CounterpartPath != nil {
		photoPath = *img.CounterpartPath
	} else {
		photoPath = img.Media.Path
	}

	photoImg, err := img.decodeImage(photoPath)
	if err != nil {
		return nil, utils.HandleError("image decoding", err)
	}

	img._photoImage = photoImg
	return img._photoImage, nil
}

func (img *EncodeMediaData) decodeImage(imagePath string) (image.Image, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open file to decode image (%s)", imagePath)
	}
	defer file.Close()

	mediaType, err := img.ContentType()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get media content type needed to decode it (%s)", imagePath)
	}

	var decodedImage image.Image

	if *mediaType == media_type.TypeHeic {
		decodedImage, _, err = image.Decode(file)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to decode HEIF image (%s)", imagePath)
		}
	} else {
		decodedImage, err = imaging.Decode(file, imaging.AutoOrientation(true))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to decode image (%s)", imagePath)
		}
	}

	return decodedImage, nil
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
