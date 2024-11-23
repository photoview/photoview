package media_type

import (
	"errors"
	"path/filepath"
	"sync"
	"testing"

	"github.com/photoview/photoview/api/test_utils"
)

func TestMagic(t *testing.T) {
	mediaPath := test_utils.PathFromAPIRoot("./scanner/test_media/real_media")

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

		{"heif.heif", mediaType("image/heic")},
		{"jpeg2000.jp2", mediaType("image/jp2")},
		{"tiff.tiff", mediaType("image/tiff")},

		{"mp4.mp4", TypeMP4},
		{"ogg.ogg", TypeOGG},
		{"mpeg.mpg", TypeMPEG},
		{"webm.webm", TypeWEBM},

		{"avi.avi", mediaType("video/x-msvideo")},
		{"mkv.mkv", mediaType("video/x-matroska")},
		{"quicktime.mov", mediaType("video/quicktime")},
		{"wmv.wmv", mediaType("video/x-ms-asf")},
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

func TestMagicNoInit(t *testing.T) {
	org := libmagic.libmagic
	libmagic.libmagic = nil
	libmagic.err = errors.New("error")
	defer func() {
		libmagic.libmagic = org
		libmagic.err = nil
	}()

	mediaPath := test_utils.PathFromAPIRoot("./scanner/test_media/real_media")
	file := filepath.Join(mediaPath, "file.pdf")

	got := GetMediaType(file)
	if want := TypeUnknown; got != want {
		t.Errorf("GetMediaType() = %v, want: %v", got, want)
	}
}
