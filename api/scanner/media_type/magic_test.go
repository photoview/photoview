package media_type

import (
	"errors"
	"flag"
	"path/filepath"
	"sync"
	"testing"

	magic "github.com/hosom/gomagic"
	"github.com/photoview/photoview/api/test_utils"
)

func init() {
	// Avoid panic with providing flags in `test_utils/integration_setup.go`.
	flag.CommandLine.Init("media_type", flag.ContinueOnError)
}

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
		{"jpg2000.jp2", mediaType("image/jp2")},
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
	t.Cleanup(func() {
		libmagic.libmagic = org
		libmagic.err = nil
	})

	mediaPath := test_utils.PathFromAPIRoot("./scanner/test_media/real_media")
	file := filepath.Join(mediaPath, "file.pdf")

	got := GetMediaType(file)
	if want := TypeUnknown; got != want {
		t.Errorf("GetMediaType() = %v, want: %v", got, want)
	}
}

func getMediaFiles() []string {
	mediaPath := test_utils.PathFromAPIRoot("./scanner/test_media/real_media")
	var files []string
	for _, f := range []string{
		"file.pdf",

		"bmp.bmp",
		"gif.gif",
		"jpeg.jpg",
		"png.png",
		"webp.webp",

		"heif.heif",
		"jpg2000.jp2",
		"tiff.tiff",

		"mp4.mp4",
		"ogg.ogg",
		"mpeg.mpg",
		"webm.webm",

		"avi.avi",
		"mkv.mkv",
		"quicktime.mov",
		"wmv.wmv",
	} {
		files = append(files, filepath.Join(mediaPath, f))
	}

	return files
}

func BenchmarkMagicAll(b *testing.B) {
	files := getMediaFiles()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m, err := magic.Open(magic.MAGIC_SYMLINK | magic.MAGIC_MIME | magic.MAGIC_ERROR | magic.MAGIC_NO_CHECK_COMPRESS | magic.MAGIC_NO_CHECK_ENCODING)
		if err != nil {
			b.Fatalf("magic.Open() error: %v", err)
		}

		_, _ = m.File(files[i%len(files)])
		m.Close()
	}
}

func BenchmarkMagicType(b *testing.B) {
	files := getMediaFiles()

	m, err := magic.Open(magic.MAGIC_SYMLINK | magic.MAGIC_MIME | magic.MAGIC_ERROR | magic.MAGIC_NO_CHECK_COMPRESS | magic.MAGIC_NO_CHECK_ENCODING)
	if err != nil {
		b.Fatalf("magic.Open() error: %v", err)
	}
	defer m.Close()

	var mu sync.Mutex

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mu.Lock()
		_, _ = m.File(files[i%len(files)])
		mu.Unlock()
	}
}
