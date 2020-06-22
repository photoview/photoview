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

	// Raw formats
	TypeDNG ImageType = "image/x-adobe-dng"
	TypeARW ImageType = "image/x-sony-arw"
	TypeSR2 ImageType = "image/x-sony-sr2"
	TypeSRF ImageType = "image/x-sony-srf"
	TypeCR2 ImageType = "image/x-canon-cr2"
	TypeCRW ImageType = "image/x-canon-crw"
	TypeERF ImageType = "image/x-epson-erf"
	TypeDCS ImageType = "image/x-kodak-dcs"
	TypeDRF ImageType = "image/x-kodak-drf"
	TypeDCR ImageType = "image/x-kodak-dcr"
	TypeK25 ImageType = "image/x-kodak-k25"
	TypeKDC ImageType = "image/x-kodak-kdc"
	TypeMRW ImageType = "image/x-minolta-mrw"
	TypeMDC ImageType = "image/x-minolta-mdc"
	TypeNEF ImageType = "image/x-nikon-nef"
	TypeNRW ImageType = "image/x-nikon-nrw"
	TypeORF ImageType = "image/x-olympus-orf"
	TypePEF ImageType = "image/x-pentax-pef"
	TypeRAF ImageType = "image/x-fuji-raf"
	TypeRAW ImageType = "image/x-panasonic-raw"
	TypeRW2 ImageType = "image/x-panasonic-rw2"
	TypeGPR ImageType = "image/x-gopro-gpr"
	Type3FR ImageType = "image/x-hasselblad-3fr"
	TypeFFF ImageType = "image/x-hasselblad-fff"
	TypeMEF ImageType = "image/x-mamiya-mef"
	TypeCap ImageType = "image/x-phaseone-cap"
	TypeIIQ ImageType = "image/x-phaseone-iiq"
	TypeMOS ImageType = "image/x-leaf-mos"
	TypeRWL ImageType = "image/x-leica-rwl"
	TypeSRW ImageType = "image/x-samsung-srw"
)

var SupportedMimetypes = [...]ImageType{
	TypeJpeg,
	TypePng,
	TypeTiff,
	TypeWebp,
	TypeBmp,

	TypeDNG,
	TypeARW,
	TypeSR2,
	TypeSRF,
	TypeCR2,
	TypeCRW,
	TypeERF,
	TypeDCS,
	TypeDRF,
	TypeDCR,
	TypeK25,
	TypeKDC,
	TypeMRW,
	TypeMDC,
	TypeNEF,
	TypeNRW,
	TypeORF,
	TypePEF,
	TypeRAF,
	TypeRAW,
	TypeRW2,
	TypeGPR,
	Type3FR,
	TypeFFF,
	TypeMEF,
	TypeCap,
	TypeIIQ,
	TypeMOS,
	TypeRWL,
	TypeSRW,
}

var WebMimetypes = [...]ImageType{
	TypeJpeg,
	TypePng,
	TypeWebp,
	TypeBmp,
}

var RawMimeTypes = [...]ImageType{
	TypeDNG,
	TypeARW,
	TypeSR2,
	TypeSRF,
	TypeCR2,
	TypeCRW,
	TypeERF,
	TypeDCS,
	TypeDRF,
	TypeDCR,
	TypeK25,
	TypeKDC,
	TypeMRW,
	TypeMDC,
	TypeNEF,
	TypeNRW,
	TypeORF,
	TypePEF,
	TypeRAF,
	TypeRAW,
	TypeRW2,
	TypeGPR,
	Type3FR,
	TypeFFF,
	TypeMEF,
	TypeCap,
	TypeIIQ,
	TypeMOS,
	TypeRWL,
	TypeSRW,
}

var fileExtensions = map[string]ImageType{
	".jpg":  TypeJpeg,
	".jpeg": TypeJpeg,
	".png":  TypePng,
	".tif":  TypeTiff,
	".tiff": TypeTiff,
	".bmp":  TypeBmp,

	// RAW formats
	".dng": TypeDNG,
	".arw": TypeARW,
	".sr2": TypeSR2,
	".srf": TypeSRF,
	".cr2": TypeCR2,
	".crw": TypeCRW,
	".erf": TypeERF,
	".dcr": TypeDCR,
	".k25": TypeK25,
	".kdc": TypeKDC,
	".mrw": TypeMRW,
	".nef": TypeNEF,
	".nrw": TypeNRW,
	".orf": TypeORF,
	".pef": TypePEF,
	".raf": TypeRAF,
	".raw": TypeRAW,
	".dcs": TypeDCS,
	".drf": TypeDRF,
	".gpr": TypeGPR,
	".3fr": Type3FR,
	".fff": TypeFFF,
}

func (imgType *ImageType) isRaw() bool {
	for _, raw_mime := range RawMimeTypes {
		if raw_mime == *imgType {
			return true
		}
	}

	return false
}

func (imgType *ImageType) isWebCompatible() bool {
	for _, web_mime := range WebMimetypes {
		if web_mime == *imgType {
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

func isPathImage(path string, cache *ScannerCache) bool {
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
