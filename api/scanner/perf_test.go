package scanner

import (
	"image"
	"image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/gographics/imagick.v3/imagick"
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

func BenchmarkMagickWand(b *testing.B) {
	dir := b.TempDir()

	imagick.Initialize()
	defer imagick.Terminate()

	for b.Loop() {
		func() {
			mw := imagick.NewMagickWand()
			defer mw.Destroy()

			if err := mw.ReadImage("./test_media/real_media/png.png"); err != nil {
				b.Fatal("read error:", err)
			}

			output := filepath.Join(dir, "test.jpg")
			defer os.Remove(output)

			if err := mw.SetFormat("JPEG"); err != nil {
				b.Fatal("set format error:", err)
			}

			if err := mw.SetImageCompressionQuality(70); err != nil {
				b.Fatal("set quality error:", err)
			}

			if err := mw.WriteImage(output); err != nil {
				b.Fatal("write error:", err)
			}
		}()
	}
}
