package media_type

import (
	"strings"
	"unique"
)

type MediaType unique.Handle[string]

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
var WebMimetypes = []MediaType{
	TypeJPEG,
	TypePNG,
	TypeWebP,
	TypeBMP,
	TypeGIF,

	TypeMP4,
	TypeMPEG,
	TypeWEBM,
	TypeOGG,
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
