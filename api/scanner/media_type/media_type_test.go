package media_type

import (
	"fmt"
	"testing"
)

type boolImage bool

const isImage boolImage = true

type boolVideo bool

const isVideo boolVideo = true

type boolWebCompatible bool

const isWebCompatible boolWebCompatible = true

type boolSupport bool

const isSupport boolSupport = true

func TestMediaTypeNoDeps(t *testing.T) {
	tests := []struct {
		mtype               MediaType
		wantIsImage         boolImage
		wantIsVideo         boolVideo
		wantIsWebCompatible boolWebCompatible
		wantIsSupport       boolSupport
	}{
		{TypeUnknown, !isImage, !isVideo, !isWebCompatible, !isSupport},
		{mediaType("application/pdf"), !isImage, !isVideo, !isWebCompatible, !isSupport},

		{mediaType("image/some-raw-type"), isImage, !isVideo, !isWebCompatible, isSupport},
		{mediaType("video/some-video-type"), !isImage, isVideo, !isWebCompatible, isSupport},

		{TypeImage, isImage, !isVideo, !isWebCompatible, isSupport},
		{TypeVideo, !isImage, isVideo, !isWebCompatible, isSupport},

		{TypeJPEG, isImage, !isVideo, isWebCompatible, isSupport},
		{TypePNG, isImage, !isVideo, isWebCompatible, isSupport},
		{TypeWebP, isImage, !isVideo, isWebCompatible, isSupport},
		{TypeBMP, isImage, !isVideo, isWebCompatible, isSupport},
		{TypeGIF, isImage, !isVideo, isWebCompatible, isSupport},

		{TypeMP4, !isImage, isVideo, isWebCompatible, isSupport},
		{TypeMPEG, !isImage, isVideo, isWebCompatible, isSupport},
		{TypeOGG, !isImage, isVideo, isWebCompatible, isSupport},
		{TypeWEBM, !isImage, isVideo, isWebCompatible, isSupport},
	}

	for _, tc := range tests {
		gotImage := tc.mtype.IsImage()
		if got, want := gotImage, bool(tc.wantIsImage); got != want {
			t.Errorf("MediaType(%q).IsImage() = %v, want: %v", tc.mtype, got, want)
		}

		gotVideo := tc.mtype.IsVideo()
		if got, want := gotVideo, bool(tc.wantIsVideo); got != want {
			t.Errorf("MediaType(%q).IsVideo() = %v, want: %v", tc.mtype, got, want)
		}

		gotWebCompatible := tc.mtype.IsWebCompatible()
		if got, want := gotWebCompatible, bool(tc.wantIsWebCompatible); got != want {
			t.Errorf("MediaType(%q).IsWebCompatible() = %v, want: %v", tc.mtype, got, want)
		}

		gotSupport := tc.mtype.IsSupported()
		if got, want := gotSupport, bool(tc.wantIsSupport); got != want {
			t.Errorf("MediaType(%q).IsSupported() = %v, want: %v", tc.mtype, got, want)
		}
	}
}

func TestMediaTypeUnknown(t *testing.T) {
	var got MediaType
	if want := TypeUnknown; got != want {
		fmt.Errorf("MediaType zero value should be TypeUnknown, which is not")
	}
}
