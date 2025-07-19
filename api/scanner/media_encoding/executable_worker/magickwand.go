package executable_worker

import (
	"fmt"

	"github.com/photoview/photoview/api/log"
	"gopkg.in/gographics/imagick.v3/imagick"
)

type MagickWand struct {
	initialized bool
}

func newMagickWand() *MagickWand {
	imagick.Initialize()

	verstr, vernum := imagick.GetVersion()

	log.Info(nil, "Found magickwand worker: "+verstr, "version", vernum)

	return &MagickWand{
		initialized: true,
	}
}

func (cli *MagickWand) Terminate() {
	cli.initialized = false
	imagick.Terminate()
}

func (cli *MagickWand) IsInstalled() bool {
	return cli != nil && cli.initialized
}

func (cli *MagickWand) EncodeJpeg(inputPath string, outputPath string, jpegQuality uint) error {
	if !cli.IsInstalled() {
		return fmt.Errorf("ImagickWand is not initialized")
	}

	wand := imagick.NewMagickWand()
	defer wand.Destroy()

	if err := wand.ReadImage(inputPath); err != nil {
		return fmt.Errorf("ImagickWand read %q error: %w", inputPath, err)
	}

	if err := wand.AutoOrientImage(); err != nil {
		return fmt.Errorf("ImagickWand auto-orient %q error: %w", inputPath, err)
	}
	// Reset EXIF orientation to 1 (top-left) since image is now properly oriented
	if err := wand.SetImageOrientation(imagick.ORIENTATION_TOP_LEFT); err != nil {
		return fmt.Errorf("ImagickWand set orientation for %q error: %w", inputPath, err)
	}

	if err := wand.SetFormat("JPEG"); err != nil {
		return fmt.Errorf("ImagickWand set JPEG format for %q error: %w", inputPath, err)
	}

	if err := wand.SetImageCompressionQuality(jpegQuality); err != nil {
		return fmt.Errorf("ImagickWand set JPEG quality %d for %q error: %w", jpegQuality, inputPath, err)
	}

	if err := wand.WriteImage(outputPath); err != nil {
		return fmt.Errorf("ImagickWand write %q error: %w", outputPath, err)
	}

	return nil
}

func (cli *MagickWand) GenerateThumbnail(inputPath string, outputPath string, width, height uint) error {
	if !cli.IsInstalled() {
		return fmt.Errorf("ImagickWand is not initialized")
	}

	wand := imagick.NewMagickWand()
	defer wand.Destroy()

	if err := wand.ReadImage(inputPath); err != nil {
		return fmt.Errorf("ImagickWand read %q error: %w", inputPath, err)
	}

	originalWidth := wand.GetImageWidth()
	originalHeight := wand.GetImageHeight()

	if err := wand.AutoOrientImage(); err != nil {
		return fmt.Errorf("ImagickWand auto-orient %q error: %w", inputPath, err)
	}
	// Reset EXIF orientation to 1 (top-left) since image is now properly oriented
	if err := wand.SetImageOrientation(imagick.ORIENTATION_TOP_LEFT); err != nil {
		return fmt.Errorf("ImagickWand set orientation for %q error: %w", inputPath, err)
	}

	// If the original image is rotated by 90 degrees, swap width and height for thumbnail generation
	if originalWidth != wand.GetImageWidth() && originalHeight != wand.GetImageHeight() {
		width, height = height, width
	}
	if err := wand.ThumbnailImage(width, height); err != nil {
		return fmt.Errorf("ImagickWand generate thumbnail for %q error: %w", inputPath, err)
	}

	if err := wand.SetFormat("JPEG"); err != nil {
		return fmt.Errorf("ImagickWand set JPEG format for %q error: %w", inputPath, err)
	}

	if err := wand.SetImageCompressionQuality(70); err != nil {
		return fmt.Errorf("ImagickWand set JPEG quality %d for %q error: %w", 70, inputPath, err)
	}

	if err := wand.WriteImage(outputPath); err != nil {
		return fmt.Errorf("ImagickWand write %q error: %w", outputPath, err)
	}

	return nil
}

func (cli *MagickWand) IdentifyDimension(inputPath string) (width, height uint, err error) {
	if !cli.IsInstalled() {
		err = fmt.Errorf("ImagickWand is not initialized")
		return
	}

	wand := imagick.NewMagickWand()
	defer wand.Destroy()

	if errRI := wand.ReadImage(inputPath); errRI != nil {
		err = fmt.Errorf("ImagickWand read %q error: %w", inputPath, errRI)
		return
	}

	width = wand.GetImageWidth()
	height = wand.GetImageHeight()

	return
}
