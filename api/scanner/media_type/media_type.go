package media_type

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/h2non/filetype"
	"github.com/photoview/photoview/api/scanner/media_encoding/executable_worker"
	"github.com/photoview/photoview/api/scanner/scanner_utils"
	"github.com/pkg/errors"
)

type MediaType string

const (
	TypeJpeg MediaType = "image/jpeg"
	TypePng  MediaType = "image/png"
	TypeTiff MediaType = "image/tiff"
	TypeWebp MediaType = "image/webp"
	TypeBmp  MediaType = "image/bmp"
	TypeHeic MediaType = "image/heic"

	// Raw formats
	TypeDNG MediaType = "image/x-adobe-dng"
	TypeARW MediaType = "image/x-sony-arw"
	TypeSR2 MediaType = "image/x-sony-sr2"
	TypeSRF MediaType = "image/x-sony-srf"
	TypeCR2 MediaType = "image/x-canon-cr2"
	TypeCRW MediaType = "image/x-canon-crw"
	TypeERF MediaType = "image/x-epson-erf"
	TypeDCS MediaType = "image/x-kodak-dcs"
	TypeDRF MediaType = "image/x-kodak-drf"
	TypeDCR MediaType = "image/x-kodak-dcr"
	TypeK25 MediaType = "image/x-kodak-k25"
	TypeKDC MediaType = "image/x-kodak-kdc"
	TypeMRW MediaType = "image/x-minolta-mrw"
	TypeMDC MediaType = "image/x-minolta-mdc"
	TypeNEF MediaType = "image/x-nikon-nef"
	TypeNRW MediaType = "image/x-nikon-nrw"
	TypeORF MediaType = "image/x-olympus-orf"
	TypePEF MediaType = "image/x-pentax-pef"
	TypeRAF MediaType = "image/x-fuji-raf"
	TypeRAW MediaType = "image/x-panasonic-raw"
	TypeRW2 MediaType = "image/x-panasonic-rw2"
	TypeGPR MediaType = "image/x-gopro-gpr"
	Type3FR MediaType = "image/x-hasselblad-3fr"
	TypeFFF MediaType = "image/x-hasselblad-fff"
	TypeMEF MediaType = "image/x-mamiya-mef"
	TypeCap MediaType = "image/x-phaseone-cap"
	TypeIIQ MediaType = "image/x-phaseone-iiq"
	TypeMOS MediaType = "image/x-leaf-mos"
	TypeRWL MediaType = "image/x-leica-rwl"
	TypeSRW MediaType = "image/x-samsung-srw"

	// Video formats
	TypeMP4  MediaType = "video/mp4"
	TypeMPEG MediaType = "video/mpeg"
	Type3GP  MediaType = "video/3gpp"
	Type3G2  MediaType = "video/3gpp2"
	TypeOGG  MediaType = "video/ogg"
	TypeWMV  MediaType = "video/x-ms-wmv"
	TypeAVI  MediaType = "video/x-msvideo"
	TypeWEBM MediaType = "video/webm"
	TypeMOV  MediaType = "video/quicktime"
	TypeTS   MediaType = "video/mp2t"
	TypeMTS  MediaType = "video/MP2T"
	TypeMKV  MediaType = "video/x-matroska"
)

var SupportedMimetypes = [...]MediaType{
	TypeJpeg,
	TypePng,
	TypeTiff,
	TypeWebp,
	TypeBmp,
	TypeHeic,
}

var WebMimetypes = [...]MediaType{
	TypeJpeg,
	TypePng,
	TypeWebp,
	TypeBmp,
}

var RawMimeTypes = [...]MediaType{
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

var VideoMimetypes = [...]MediaType{
	TypeMP4,
	TypeMPEG,
	Type3GP,
	Type3G2,
	TypeOGG,
	TypeWMV,
	TypeAVI,
	TypeWEBM,
	TypeMOV,
	TypeTS,
	TypeMTS,
	TypeMKV,
}

// WebVideoMimetypes are video types that can be played directly in the browser without transcoding
var WebVideoMimetypes = [...]MediaType{
	TypeMP4,
	TypeMPEG,
	TypeWEBM,
	TypeOGG,
}

var fileExtensions = map[string]MediaType{
	".jpg":  TypeJpeg,
	".jpeg": TypeJpeg,
	".png":  TypePng,
	".tif":  TypeTiff,
	".tiff": TypeTiff,
	".bmp":  TypeBmp,
	".heic": TypeHeic,

	// RAW formats
	".dng": TypeDNG,
	".arw": TypeARW,
	".sr2": TypeSR2,
	".srf": TypeSRF,
	".srw": TypeSRW,
	".cr2": TypeCR2,
	".crw": TypeCRW,
	".erf": TypeERF,
	".dcr": TypeDCR,
	".k25": TypeK25,
	".kdc": TypeKDC,
	".mrw": TypeMRW,
	".nef": TypeNEF,
	".nrw": TypeNRW,
	".mdc": TypeMDC,
	".mef": TypeMEF,
	".orf": TypeORF,
	".pef": TypePEF,
	".raf": TypeRAF,
	".raw": TypeRAW,
	".rw2": TypeRW2,
	".dcs": TypeDCS,
	".drf": TypeDRF,
	".gpr": TypeGPR,
	".3fr": Type3FR,
	".fff": TypeFFF,
	".cap": TypeCap,
	".iiq": TypeIIQ,
	".mos": TypeMOS,
	".rwl": TypeRWL,

	// Video formats
	".mp4":  TypeMP4,
	".m4v":  TypeMP4,
	".mpeg": TypeMPEG,
	".3gp":  Type3GP,
	".3g2":  Type3G2,
	".ogv":  TypeOGG,
	".wmv":  TypeWMV,
	".avi":  TypeAVI,
	".webm": TypeWEBM,
	".mov":  TypeMOV,
	".qt":   TypeMOV,
	".ts":   TypeTS,
	".m2ts": TypeMTS,
	".mts":  TypeMTS,
	".mkv":  TypeMKV,
}

func (imgType *MediaType) IsRaw() bool {
	for _, raw_mime := range RawMimeTypes {
		if raw_mime == *imgType {
			return true
		}
	}

	return false
}

func (imgType *MediaType) IsWebCompatible() bool {
	for _, web_mime := range WebMimetypes {
		if web_mime == *imgType {
			return true
		}
	}

	for _, web_mime := range WebVideoMimetypes {
		if web_mime == *imgType {
			return true
		}
	}

	return false
}

func (imgType *MediaType) IsVideo() bool {
	for _, video_mime := range VideoMimetypes {
		if video_mime == *imgType {
			return true
		}
	}

	return false
}

func (imgType *MediaType) IsBasicTypeSupported() bool {
	for _, img_mime := range SupportedMimetypes {
		if img_mime == *imgType {
			return true
		}
	}

	return false
}

// IsSupported determines if the given type can be processed
func (imgType *MediaType) IsSupported() bool {
	if imgType.IsBasicTypeSupported() {
		return true
	}

	if executable_worker.DarktableCli.IsInstalled() && imgType.IsRaw() {
		return true
	}

	if executable_worker.FfmpegCli.IsInstalled() && imgType.IsVideo() {
		return true
	}

	return false
}

func GetExtensionMediaType(ext string) (MediaType, bool) {
	result, found := fileExtensions[strings.ToLower(ext)]
	return result, found
}

func GetMediaType(path string) (*MediaType, error) {
	ext := filepath.Ext(path)

	fileExtType, found := GetExtensionMediaType(ext)

	if found {
		if fileExtType.IsSupported() {
			return &fileExtType, nil
		} else {
			return nil, nil
		}
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

	imgType := MediaType(_imgType.MIME.Value)
	if imgType.IsSupported() {
		return &imgType, nil
	}

	return nil, nil
}

func (mediaType MediaType) FileExtensions() []string {
	var extensions []string

	for ext, extType := range fileExtensions {
		if extType == mediaType {
			extensions = append(extensions, ext)
			extensions = append(extensions, strings.ToUpper(ext))
		}
	}

	return extensions
}

func RawCounterpart(imagePath string) *string {
	pathWithoutExt := strings.TrimSuffix(imagePath, path.Ext(imagePath))

	for _, rawType := range RawMimeTypes {
		for _, ext := range rawType.FileExtensions() {
			testPath := pathWithoutExt + ext
			if scanner_utils.FileExists(testPath) {
				return &testPath
			}
		}
	}

	return nil
}
