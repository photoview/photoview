package media_type

import (
	"strings"
	"unique"

	"github.com/photoview/photoview/api/log"
	"github.com/photoview/photoview/api/scanner/externaltools/exif"
	"github.com/spf13/afero"
)

type MediaType unique.Handle[string]

// GetMediaType returns a media type of file `f`.
// This function is thread-safe.
func GetMediaType(fs afero.Fs, f string) MediaType {
	mime, err := exif.MIMEType(fs, f)
	if err != nil {
		log.Warn(nil, "GetMediaType() error.", "error", err, "file", f)
		return TypeUnknown
	}

	return mediaType(mime)
}

func mediaType(mime string) MediaType {
	return MediaType(unique.Make(mime))
}

var (
	TypeUnknown MediaType

	TypeImage = mediaType("image/")
	TypeVideo = mediaType("video/")

	// Web Image formats
	TypeJPEG = mediaType("image/jpeg")
	TypePNG  = mediaType("image/png")
	TypeWebP = mediaType("image/webp")
	TypeBMP  = mediaType("image/bmp")
	TypeGIF  = mediaType("image/gif")

	// Web Video formats
	TypeMP4  = mediaType("video/mp4")
	TypeMPEG = mediaType("video/mpeg")
	TypeOGG  = mediaType("video/ogg")
	TypeWEBM = mediaType("video/webm")
)

var webImageMimetypes = arrayToSet([]MediaType{
	TypeJPEG,
	TypePNG,
	TypeWebP,
	TypeBMP,
	TypeGIF,
})

var webVideoMimetypes = arrayToSet([]MediaType{
	TypeMP4,
	TypeMPEG,
	TypeWEBM,
	TypeOGG,
})

// Legacy function. Should be removed.
var WebMimetypes = []string{
	TypeJPEG.String(),
	TypePNG.String(),
	TypeWebP.String(),
	TypeBMP.String(),
	TypeGIF.String(),

	TypeMP4.String(),
	TypeMPEG.String(),
	TypeWEBM.String(),
	TypeOGG.String(),
}

func arrayToSet[T comparable](array []T) map[T]struct{} {
	ret := make(map[T]struct{})
	for _, item := range array {
		ret[item] = struct{}{}
	}
	return ret
}

// IsWebCompatible returns true if the media type is compatible with the browser.
func (t MediaType) IsWebCompatible() bool {
	if t == TypeUnknown {
		return false
	}

	if _, ok := webImageMimetypes[t]; ok {
		return true
	}

	if _, ok := webVideoMimetypes[t]; ok {
		return true
	}

	return false
}

// IsImage returns true if the media type is image type.
func (t MediaType) IsImage() bool {
	if t == TypeUnknown {
		return false
	}

	return strings.HasPrefix(t.String(), TypeImage.String())
}

// IsVideo returns true if the media type is video type.
func (t MediaType) IsVideo() bool {
	if t == TypeUnknown {
		return false
	}

	return strings.HasPrefix(t.String(), TypeVideo.String())
}

// IsSupported returns true if the media type can be processed.
func (t MediaType) IsSupported() bool {
	if t == TypeUnknown {
		return false
	}

	return t.IsImage() || t.IsVideo()
}

func (t MediaType) String() string {
	if t == TypeUnknown {
		return "unknown"
	}

	return unique.Handle[string](t).Value()
}
