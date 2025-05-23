package scanner

import (
	"image"
	"image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/kkovaletp/photoview/api/scanner/media_encoding/executable_worker"
)

func BenchmarkStdlib(b *testing.B) {
	dir := b.TempDir()

	for b.Loop() {
		func() {
			input, err := os.Open("./test_media/real_media/png.png")
			if err != nil {
				b.Fatal("open error:", err)
			}
			defer input.Close()

			img, _, err := image.Decode(input)
			if err != nil {
				b.Fatal("decode error:", err)
			}

			outfile := filepath.Join(dir, "test.jpg")
			defer os.Remove(outfile)

			output, err := os.Create(outfile)
			if err != nil {
				b.Fatal("create error:", err)
			}
			defer output.Close()

			if err := jpeg.Encode(output, img, &jpeg.Options{Quality: 70}); err != nil {
				b.Fatal("encode error:", err)
			}
		}()
	}
}

func BenchmarkMagickCLI(b *testing.B) {
	dir := b.TempDir()

	for b.Loop() {
		func() {
			output := filepath.Join(dir, "test.jpg")
			defer os.Remove(output)

			if err := executable_worker.Magick.EncodeJpeg("./test_media/real_media/png.png", output, 70); err != nil {
				b.Fatal("encode jpeg error:", err)
			}
		}()
	}
}
