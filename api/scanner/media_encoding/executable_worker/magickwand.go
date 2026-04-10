package executable_worker

import (
	"fmt"

	"github.com/gographics/gmagick"

	"github.com/photoview/photoview/api/log"
)

type MagickWand struct {
	initialized bool
}

func newMagickWand() *MagickWand {
	gmagick.Initialize()

	verstr, vernum := gmagick.GetVersion()

	log.Info(nil, "Found magickwand worker: "+verstr, "version", vernum)

	return &MagickWand{
		initialized: true,
	}
}

func (cli *MagickWand) Terminate() {
	cli.initialized = false
	gmagick.Terminate()
}

func (cli *MagickWand) IsInstalled() bool {
	return cli != nil && cli.initialized
}

func (cli *MagickWand) GenerateThumbnail(inputPath string, outputPath string, width, height uint) error {
	wand, err := cli.createWandFromFile(inputPath)
	if err != nil {
		return err
	}
	defer wand.Destroy()

	if err := wand.ResizeImage(width, height, gmagick.FILTER_LANCZOS, 1); err != nil {
		return fmt.Errorf("ImagickWand generate thumbnail for %q error: %w", inputPath, err)
	}

	if err := wand.SetFormat("JPEG"); err != nil {
		return fmt.Errorf("ImagickWand set JPEG format for %q error: %w", inputPath, err)
	}

	if err := wand.SetCompressionQuality(70); err != nil {
		return fmt.Errorf("ImagickWand set JPEG quality %d for %q error: %w", 70, inputPath, err)
	}

	if err := wand.WriteImage(outputPath); err != nil {
		return fmt.Errorf("ImagickWand write %q error: %w", outputPath, err)
	}

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

	return
}

func (cli *MagickWand) createWandFromFile(inputPath string) (*gmagick.MagickWand, error) {
	if !cli.IsInstalled() {
		return nil, fmt.Errorf("ImagickWand is not initialized")
	}

	wand := gmagick.NewMagickWand()

	if err := wand.ReadImage(inputPath); err != nil {
		return nil, fmt.Errorf("ImagickWand read %q error: %w", inputPath, err)
	}

	if err := wand.AutoOrientImage(gmagick.ORIENTATION_UNDEFINED); err != nil {
		return nil, fmt.Errorf("ImagickWand auto-orient %q error: %w", inputPath, err)
	}

	// Reset EXIF orientation to 1 (top-left) since image is now properly oriented
	if err := wand.SetImageOrientation(gmagick.ORIENTATION_TOP_LEFT); err != nil {
		return nil, fmt.Errorf("ImagickWand set orientation for %q error: %w", inputPath, err)
	}

	return wand, nil
}
