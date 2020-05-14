package scanner

import (
	"database/sql"
	"image"
	"image/jpeg"
	"log"
	"os"

	"github.com/disintegration/imaging"
	"github.com/h2non/filetype"
	cr2Decoder "github.com/nf/cr2"
	"github.com/viktorstrate/photoview/api/graphql/models"
	"github.com/viktorstrate/photoview/api/utils"
)

func encodeImageJPEG(photoPath string, photoImage image.Image, jpegOptions *jpeg.Options) error {
	photo_file, err := os.Create(photoPath)
	if err != nil {
		log.Printf("ERROR: Could not create file: %s\n", photoPath)
		return err
	}
	defer photo_file.Close()

	err = jpeg.Encode(photo_file, photoImage, jpegOptions)
	if err != nil {
		return err
	}

	return nil
}

// EncodeImageData is used to easily decode image data, with a cache so expensive operations are not repeated
type EncodeImageData struct {
	photo           *models.Photo
	_photoImage     image.Image
	_thumbnailImage image.Image
	_contentType    *ImageType
}

// ContentType reads the image to determine its content type
func (img *EncodeImageData) ContentType() (*ImageType, error) {
	if img._contentType != nil {
		return img._contentType, nil
	}

	file, err := os.Open(img.photo.Path)
	if err != nil {
		ScannerError("Could not open file %s: %s\n", img.photo.Path, err)
		return nil, err
	}
	defer file.Close()

	head := make([]byte, 261)
	if _, err := file.Read(head); err != nil {
		ScannerError("Could not read photo %s: %s\n", img.photo.Path, err)
		return nil, err
	}

	_imgType, err := filetype.Image(head)
	if err != nil {
		return nil, err
	}

	imgType := ImageType(_imgType.MIME.Value)
	img._contentType = &imgType
	return img._contentType, nil
}

// PhotoImage reads and decodes the image file and saves it in a cache so the photo in only decoded once
func (img *EncodeImageData) PhotoImage(tx *sql.Tx) (image.Image, error) {
	if img._photoImage != nil {
		return img._photoImage, nil
	}

	photoFile, err := os.Open(img.photo.Path)
	if err != nil {
		return nil, err
	}
	defer photoFile.Close()

	var photoImg image.Image
	contentType, err := img.ContentType()
	if err != nil {
		return nil, err
	}

	if contentType != nil && *contentType == "image/x-canon-cr2" {
		photoImg, err = cr2Decoder.Decode(photoFile)
		if err != nil {
			return nil, utils.HandleError("cr2 raw image decoding", err)
		}
	} else {
		photoImg, _, err = image.Decode(photoFile)
		if err != nil {
			return nil, utils.HandleError("image decoding", err)
		}
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

// ThumbnailImage downsizes the image and returns it
func (img *EncodeImageData) ThumbnailImage(tx *sql.Tx) (image.Image, error) {
	photoImage, err := img.PhotoImage(tx)
	if err != nil {
		return nil, err
	}

	dimensions := photoImage.Bounds().Max
	aspect := float64(dimensions.X) / float64(dimensions.Y)

	var width, height int

	if aspect > 1 {
		width = 1024
		height = int(1024 / aspect)
	} else {
		width = int(1024 * aspect)
		height = 1024
	}

	thumbImage := imaging.Thumbnail(photoImage, width, height, imaging.NearestNeighbor)
	img._thumbnailImage = thumbImage

	return img._thumbnailImage, nil
}
