package scanner

import (
	"database/sql"
	"image"
	"image/jpeg"
	"os"

	"github.com/disintegration/imaging"
	"github.com/pkg/errors"
	"github.com/viktorstrate/photoview/api/graphql/models"
	"github.com/viktorstrate/photoview/api/utils"
)

type PhotoDimensions struct {
	Width  int
	Height int
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

// EncodeImageData is used to easily decode image data, with a cache so expensive operations are not repeated
type EncodeImageData struct {
	photo           *models.Photo
	_photoImage     image.Image
	_thumbnailImage image.Image
	_contentType    *MediaType
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
func (img *EncodeImageData) ContentType() (*MediaType, error) {
	if img._contentType != nil {
		return img._contentType, nil
	}

	imgType, err := getImageType(img.photo.Path)
	if err != nil {
		return nil, err
	}

	img._contentType = imgType
	return imgType, nil
}

func (img *EncodeImageData) EncodeHighRes(tx *sql.Tx, outputPath string) error {
	contentType, err := img.ContentType()
	if err != nil {
		return err
	}

	if !contentType.isSupported() {
		return errors.New("could not convert photo as file format is not supported")
	}

	if contentType.isRaw() {
		if DarktableCli.IsInstalled() {
			err := DarktableCli.EncodeJpeg(img.photo.Path, outputPath, 70)
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
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return nil, err
	}
	defer inputFile.Close()

	inputImage, _, err := image.Decode(inputFile)
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
func (img *EncodeImageData) photoImage(tx *sql.Tx) (image.Image, error) {
	if img._photoImage != nil {
		return img._photoImage, nil
	}

	photoFile, err := os.Open(img.photo.Path)
	if err != nil {
		return nil, err
	}
	defer photoFile.Close()

	photoImg, _, err := image.Decode(photoFile)
	if err != nil {
		return nil, utils.HandleError("image decoding", err)
	}

	// Get orientation from exif data
	row := tx.QueryRow("SELECT photo_exif.orientation FROM photo JOIN photo_exif WHERE photo.exif_id = photo_exif.exif_id AND photo.photo_id = ?", img.photo.PhotoID)
	var orientation *int
	if err = row.Scan(&orientation); err != nil {
		// If not found use default orientation (not rotate)
		if err == sql.ErrNoRows {
			orientation = nil
		} else {
			return nil, err
		}
	}

	if orientation == nil {
		defaultOrientation := 0
		orientation = &defaultOrientation
	}

	switch *orientation {
	case 2:
		photoImg = imaging.FlipH(photoImg)
		break
	case 3:
		photoImg = imaging.Rotate180(photoImg)
		break
	case 4:
		photoImg = imaging.FlipV(photoImg)
		break
	case 5:
		photoImg = imaging.Transpose(photoImg)
		break
	case 6:
		photoImg = imaging.Rotate270(photoImg)
		break
	case 7:
		photoImg = imaging.Transverse(photoImg)
		break
	case 8:
		photoImg = imaging.Rotate90(photoImg)
		break
	default:
		break
	}

	img._photoImage = photoImg
	return img._photoImage, nil
}
