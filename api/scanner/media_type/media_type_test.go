package media_type

import (
	"path/filepath"
	"sync"
	"testing"

	"github.com/photoview/photoview/api/test_utils"
)

func TestMain(m *testing.M) {
	test_utils.IntegrationTestRun(m)
}

func TestGetMediaType(t *testing.T) {
	mediaPath := test_utils.PathFromAPIRoot("scanner", "test_media", "real_media")

	tests := []struct {
		filepath string
		filetype MediaType
	}{
		{"file.pdf", mediaType("application/pdf")},

		{"bmp.bmp", TypeBMP},
		{"gif.gif", TypeGIF},
		{"jpeg.jpg", TypeJPEG},
		{"png.png", TypePNG},
		{"webp.webp", TypeWebP},

		{"heif.heif", mediaType("image/heif")},
		{"jpg2000.jp2", mediaType("image/jp2")},
		{"tiff.tiff", mediaType("image/tiff")},
		{"cr3.cr3", mediaType("image/x-canon-cr3")},

		{"mp4.mp4", TypeMP4},
		{"ogg.ogg", TypeOGG},
		{"mpeg.mpg", TypeMPEG},
		{"webm.webm", TypeWEBM},

		{"avi.avi", mediaType("video/vnd.avi")},
		{"mkv.mkv", mediaType("video/x-matroska")},
		{"quicktime.mov", mediaType("video/quicktime")},
		{"wmv.wmv", mediaType("video/x-ms-wmv")},
	}

	var wg sync.WaitGroup
	defer wg.Wait()

	for _, tc := range tests {
		wg.Add(1)
		input, want := tc.filepath, tc.filetype

		go func() {
			defer wg.Done()

			path := filepath.Join(mediaPath, input)

			got := GetMediaType(path)
			if got != want {
				t.Errorf("magic.Type(%q) = %v, want: %v", path, got, want)
			}
		}()
	}
}

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
		// Unknown and unsupported types
		{TypeUnknown, !isImage, !isVideo, !isWebCompatible, !isSupport},
		{mediaType("application/pdf"), !isImage, !isVideo, !isWebCompatible, !isSupport},

		// Raw media types
		{mediaType("image/some-raw-type"), isImage, !isVideo, !isWebCompatible, isSupport},
		{mediaType("video/some-video-type"), !isImage, isVideo, !isWebCompatible, isSupport},

		// Generic types
		{TypeImage, isImage, !isVideo, !isWebCompatible, isSupport},
		{TypeVideo, !isImage, isVideo, !isWebCompatible, isSupport},

		// Web-compatible image types
		{TypeJPEG, isImage, !isVideo, isWebCompatible, isSupport},
		{TypePNG, isImage, !isVideo, isWebCompatible, isSupport},
		{TypeWebP, isImage, !isVideo, isWebCompatible, isSupport},
		{TypeBMP, isImage, !isVideo, isWebCompatible, isSupport},
		{TypeGIF, isImage, !isVideo, isWebCompatible, isSupport},

		// Web-compatible video types
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
		t.Errorf("MediaType zero value should be TypeUnknown, which is not")
	}
}
