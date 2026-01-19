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
	wand, err := cli.createWandFromFile(inputPath)
	if err != nil {
		return err
	}
	defer wand.Destroy()

	if err := wand.SetFormat("JPEG"); err != nil {
		return fmt.Errorf("ImagickWand set JPEG format for %q error: %w", inputPath, err)
	}

	if err := wand.SetImageCompressionQuality(jpegQuality); err != nil {
		return fmt.Errorf("ImagickWand set JPEG quality %d for %q error: %w", jpegQuality, inputPath, err)
	}

	if err := wand.WriteImage(outputPath); err != nil {
		return fmt.Errorf("ImagickWand write %q error: %w", outputPath, err)
	}

	fmt.Printf("Encoded JPEG %q to %q\n", inputPath, outputPath)

	return nil
}

func (cli *MagickWand) GenerateThumbnail(inputPath string, outputPath string, width, height uint) error {
	wand, err := cli.createWandFromFile(inputPath)
	if err != nil {
		return err
	}
	defer wand.Destroy()

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

	fmt.Printf("Generated thumbnail %q to %q\n", inputPath, outputPath)

	return nil
}

func (cli *MagickWand) IdentifyDimension(inputPath string) (width, height uint, reterr error) {
	wand, err := cli.createWandFromFile(inputPath)
	if err != nil {
		reterr = err
		return
	}
	defer wand.Destroy()

	width = wand.GetImageWidth()
	height = wand.GetImageHeight()

	fmt.Printf("Identified dimensions for %q: %dx%d\n", inputPath, width, height)

	return
}

func (cli *MagickWand) createWandFromFile(inputPath string) (*imagick.MagickWand, error) {
	if !cli.IsInstalled() {
		return nil, fmt.Errorf("ImagickWand is not initialized")
	}

	wand := imagick.NewMagickWand()

	if err := wand.ReadImage(inputPath); err != nil {
		return nil, fmt.Errorf("ImagickWand read %q error: %w", inputPath, err)
	}

	if err := wand.AutoOrientImage(); err != nil {
		return nil, fmt.Errorf("ImagickWand auto-orient %q error: %w", inputPath, err)
	}

	// Reset EXIF orientation to 1 (top-left) since image is now properly oriented
	if err := wand.SetImageOrientation(imagick.ORIENTATION_TOP_LEFT); err != nil {
		return nil, fmt.Errorf("ImagickWand set orientation for %q error: %w", inputPath, err)
	}

	return wand, nil
}
