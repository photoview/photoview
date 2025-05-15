package media_utils

import (
	"fmt"
	"image"

	"github.com/photoview/photoview/api/scanner/media_encoding/executable_worker"
)

type PhotoDimensions struct {
	Width  int
	Height int
}

func GetPhotoDimensions(imagePath string) (*PhotoDimensions, error) {
	ret, err := executable_worker.Magick.IdentifyDimension(imagePath)
	if err != nil {
		return nil, fmt.Errorf("identify dimension %q error: %w", imagePath, err)
	}

	return &PhotoDimensions{
		Width:  ret.Width,
		Height: ret.Height,
	}, nil
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

	if width > dimensions.Width {
		width = dimensions.Width
		height = dimensions.Height
	}

	return PhotoDimensions{
		Width:  width,
		Height: height,
	}
}
