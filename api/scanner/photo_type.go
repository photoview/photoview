package scanner

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/h2non/filetype"
	"github.com/pkg/errors"
)

type ImageType string

const (
	TypeJpeg ImageType = "image/jpeg"
	TypePng  ImageType = "image/png"
	TypeTiff ImageType = "image/tiff"
	TypeWebp ImageType = "image/webp"
	TypeBmp  ImageType = "image/bmp"
	TypeCr2  ImageType = "image/x-canon-cr2"
)

var SupportedMimetypes = [...]ImageType{
	TypeJpeg,
	TypePng,
	TypeTiff,
	TypeWebp,
	TypeBmp,
	TypeCr2,
}

var WebMimetypes = [...]ImageType{
	TypeJpeg,
	TypePng,
	TypeWebp,
	TypeBmp,
}

var RawMimeTypes = [...]ImageType{
	TypeCr2,
}

var fileExtensions = map[string]ImageType{
	".jpg":  TypeJpeg,
	".jpeg": TypeJpeg,
	".png":  TypePng,
	".tif":  TypeTiff,
	".tiff": TypeTiff,
	".bmp":  TypeBmp,
	".cr2":  TypeCr2,
}

func (imgType *ImageType) isRaw() bool {
	for _, raw_mime := range RawMimeTypes {
		if raw_mime == *imgType {
			return true
		}
	}

	return false
}

func (imgType *ImageType) isSupported() bool {
	for _, supported_mime := range SupportedMimetypes {
		if supported_mime == *imgType {
			return true
		}
	}

	return false
}

func getImageType(path string) (*ImageType, error) {

	ext := filepath.Ext(path)

	fileExtType := fileExtensions[strings.ToLower(ext)]

	if fileExtType.isSupported() {
		return &fileExtType, nil
	}

	// If extension was not recognized try to read file header
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "could not open file to determine content-type %s", path)
	}
	defer file.Close()

	head := make([]byte, 261)
	if _, err := file.Read(head); err != nil {
		if err == io.EOF {
			return nil, nil
		}

		return nil, errors.Wrapf(err, "could not read file to determine content-type: %s", path)
	}

	_imgType, err := filetype.Image(head)
	if err != nil {
		return nil, nil
	}

	imgType := ImageType(_imgType.MIME.Value)
	if imgType.isSupported() {
		return &imgType, nil
	}

	return nil, nil
}

func isPathImage(path string, cache *scanner_cache) bool {
	if cache.get_photo_type(path) != nil {
		return true
	}

	imageType, err := getImageType(path)
	if err != nil {
		ScannerError("%s (%s)", err, path)
		return false
	}

	if imageType != nil {
		// Make sure file isn't empty
		fileStats, err := os.Stat(path)
		if err != nil || fileStats.Size() == 0 {
			return false
		}

		cache.insert_photo_type(path, *imageType)
		return true
	}

	log.Printf("File is not a supported image %s\n", path)
	return false
}