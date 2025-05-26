package executable_worker

import (
	"fmt"

	"github.com/photoview/photoview/api/log"
	"gopkg.in/gographics/imagick.v3/imagick"
)

type MagickWand struct {
}

func newMagickWand() *MagickWand {
	imagick.Initialize()

	verstr, vernum := imagick.GetVersion()

	log.Info("Found magickwand worker: "+verstr, "version", vernum)

	return &MagickWand{}
}

func (cli *MagickWand) IsInstalled() bool {
	return true
}

func (cli *MagickWand) EncodeJpeg(inputPath string, outputPath string, jpegQuality int) error {
	wand := imagick.NewMagickWand()
	defer wand.Destroy()

	if err := wand.ReadImage(inputPath); err != nil {
		return fmt.Errorf("ImagickWand read %q error: %w", inputPath, err)
	}

	if err := wand.SetFormat("JPEG"); err != nil {
		return fmt.Errorf("ImagickWand set JPEG format for %q error: %w", inputPath, err)
	}

	if err := wand.SetImageCompressionQuality(70); err != nil {
		return fmt.Errorf("ImagickWand set JEPG quality %d for %q error: %w", jpegQuality, inputPath, err)
	}

	if err := wand.WriteImage(outputPath); err != nil {
		return fmt.Errorf("ImagickWand write %q error: %w", outputPath, err)
	}

	return nil
}

func (cli *MagickWand) GenerateThumbnail(inputPath string, outputPath string, width, height int) error {
	wand := imagick.NewMagickWand()
	defer wand.Destroy()

	if err := wand.ReadImage(inputPath); err != nil {
		return fmt.Errorf("ImagickWand read %q error: %w", inputPath, err)
	}

	if err := wand.ThumbnailImage(uint(width), uint(height)); err != nil {
		return fmt.Errorf("ImagickWand generate thumbnail for %q error: %w", inputPath, err)
	}

	if err := wand.SetFormat("JPEG"); err != nil {
		return fmt.Errorf("ImagickWand set JPEG format for %q error: %w", inputPath, err)
	}

	if err := wand.SetImageCompressionQuality(70); err != nil {
		return fmt.Errorf("ImagickWand set JEPG quality %d for %q error: %w", 70, inputPath, err)
	}

	if err := wand.WriteImage(outputPath); err != nil {
		return fmt.Errorf("ImagickWand write %q error: %w", outputPath, err)
	}

	return nil
}

func (cli *MagickWand) IdentifyDimension(inputPath string) (width, height int, err error) {
	wand := imagick.NewMagickWand()
	defer wand.Destroy()

	if e := wand.ReadImage(inputPath); e != nil {
		err = fmt.Errorf("ImagickWand read %q error: %w", inputPath, e)
		return
	}

	width = int(wand.GetImageWidth())
	height = int(wand.GetImageHeight())

	return
}
