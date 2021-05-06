package media_utils

import (
	"image"
	"os"
)

type PhotoDimensions struct {
	Width  int
	Height int
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
