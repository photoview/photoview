package scanner

import (
	"image"
	"image/jpeg"
	"os"

	"github.com/disintegration/imaging"
	"github.com/pkg/errors"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/utils"
	"gopkg.in/vansante/go-ffprobe.v2"
	"gorm.io/gorm"
)

type PhotoDimensions struct {
	Width  int
	Height int
}

func DecodeImage(imagePath string) (image.Image, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open file to decode image (%s)", imagePath)
	}
	defer file.Close()

	image, err := imaging.Decode(file, imaging.AutoOrientation(true))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode image (%s)", imagePath)
	}

	return image, nil
}

func PhotoDimensionsFromRect(rect image.Rectangle) PhotoDimensions {
	return PhotoDimensions{
		Width:  rect.Bounds().Max.X,
		Height: rect.Bounds().Max.Y,
	}
}

func (dimensions *PhotoDimensions) ThumbnailScale() PhotoDimensions {
	aspect := float64(dimensions.Width) / float64(dimensions.Height)

	var width, height int

	if aspect > 1 {
		width = 1024
		height = int(1024 / aspect)
	} else {
		width = int(1024 * aspect)
		height = 1024
	}

	return PhotoDimensions{
		Width:  width,
		Height: height,
	}
}

// EncodeMediaData is used to easily decode media data, with a cache so expensive operations are not repeated
type EncodeMediaData struct {
	media           *models.Media
	_photoImage     image.Image
	_thumbnailImage image.Image
	_contentType    *MediaType
	_videoMetadata  *ffprobe.ProbeData
}

func EncodeImageJPEG(image image.Image, outputPath string, jpegQuality int) error {
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

func GetPhotoDimensions(imagePath string) (*PhotoDimensions, error) {
	photoFile, err := os.Open(imagePath)
	if err != nil {
		return nil, err
	}
	defer photoFile.Close()

	config, _, err := image.DecodeConfig(photoFile)
	if err != nil {
		return nil, err
	}

	return &PhotoDimensions{
		Width:  config.Width,
		Height: config.Height,
	}, nil
}

// ContentType reads the image to determine its content type
func (img *EncodeMediaData) ContentType() (*MediaType, error) {
	if img._contentType != nil {
		return img._contentType, nil
	}

	imgType, err := getMediaType(img.media.Path)
	if err != nil {
		return nil, err
	}

	img._contentType = imgType
	return imgType, nil
}

func (img *EncodeMediaData) EncodeHighRes(tx *gorm.DB, outputPath string) error {
	contentType, err := img.ContentType()
	if err != nil {
		return err
	}

	if !contentType.isSupported() {
		return errors.New("could not convert photo as file format is not supported")
	}

	// Use darktable if there is no counterpart JPEG file to use instead
	if contentType.isRaw() && img.media.CounterpartPath == nil {
		if DarktableCli.IsInstalled() {
			err := DarktableCli.EncodeJpeg(img.media.Path, outputPath, 70)
			if err != nil {
				return err
			}
		} else {
			return errors.New("could not convert photo as no RAW converter was found")
		}
	} else {
		image, err := img.photoImage(tx)
		if err != nil {
			return err
		}

		EncodeImageJPEG(image, outputPath, 70)
	}

	return nil
}

func EncodeThumbnail(inputPath string, outputPath string) (*PhotoDimensions, error) {
	inputImage, err := DecodeImage(inputPath)
	if err != nil {
		return nil, err
	}

	dimensions := PhotoDimensionsFromRect(inputImage.Bounds())
	dimensions = dimensions.ThumbnailScale()

	thumbImage := imaging.Resize(inputImage, dimensions.Width, dimensions.Height, imaging.NearestNeighbor)
	if err = EncodeImageJPEG(thumbImage, outputPath, 60); err != nil {
		return nil, err
	}

	return &dimensions, nil
}

// PhotoImage reads and decodes the image file and saves it in a cache so the photo in only decoded once
func (img *EncodeMediaData) photoImage(tx *gorm.DB) (image.Image, error) {
	if img._photoImage != nil {
		return img._photoImage, nil
	}

	var photoPath string
	if img.media.CounterpartPath != nil {
		photoPath = *img.media.CounterpartPath
	} else {
		photoPath = img.media.Path
	}

	photoImg, err := DecodeImage(photoPath)
	if err != nil {
		return nil, utils.HandleError("image decoding", err)
	}

	img._photoImage = photoImg
	return img._photoImage, nil
}
