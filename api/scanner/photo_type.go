package scanner

import (
	"io"
	"log"
	"os"

	"github.com/h2non/filetype"
)

type FileType string

const (
	TypeJpeg FileType = "image/jpeg"
	TypePng  FileType = "image/png"
	TypeTiff FileType = "image/tiff"
	TypeWebp FileType = "image/webp"
	TypeCr2  FileType = "image/x-canon-cr2"
	TypeBmp  FileType = "image/bmp"
)

var SupportedMimetypes = [...]FileType{
	TypeJpeg,
	TypePng,
	TypeTiff,
	TypeWebp,
	TypeBmp,
	TypeCr2,
}

var WebMimetypes = [...]FileType{
	TypeJpeg,
	TypePng,
	TypeWebp,
	TypeBmp,
}

func isPathImage(path string, cache *scanner_cache) bool {
	if cache.get_photo_type(path) != nil {
		return true
	}
	file, err := os.Open(path)
	if err != nil {
		ScannerError("Could not open file %s: %s\n", path, err)
		return false
	}
	defer file.Close()

	head := make([]byte, 261)
	if _, err := file.Read(head); err != nil {
		if err == io.EOF {
			return false
		}

		ScannerError("Could not read file %s: %s\n", path, err)
		return false
	}

	imgType, err := filetype.Image(head)
	if err != nil {
		return false
	}

	for _, supported_mime := range SupportedMimetypes {
		if supported_mime == FileType(imgType.MIME.Value) {
			cache.insert_photo_type(path, supported_mime)
			return true
		}
	}

	log.Printf("Unsupported image %s of type %s\n", path, imgType.MIME.Value)
	return false
}
