package exif

import (
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/pkg/errors"
	"github.com/xor-gate/goexif2/exif"
	"github.com/xor-gate/goexif2/mknote"
)

// internalExifParser is an exif parser that parses the media without the use of external tools
type internalExifParser struct{}

func (p *internalExifParser) ParseExif(media *models.Media) (returnExif *models.MediaEXIF, returnErr error) {
	photoFile, err := os.Open(media.Path)
	if err != nil {
		return nil, err
	}

	exif.RegisterParsers(mknote.All...)

	// Recover if exif.Decode panics
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Recovered from panic: Exif decoding: %s\n", err)
			returnErr = errors.New(fmt.Sprintf("Exif decoding panicked: %s\n", err))
		}
	}()

	exifTags, err := exif.Decode(photoFile)
	if err != nil {
		return nil, errors.Wrap(err, "Could not decode EXIF")
	}

	newExif := models.MediaEXIF{}

	model, err := p.readStringTag(exifTags, exif.Model, media)
	if err == nil {
		newExif.Camera = model
	}

	maker, err := p.readStringTag(exifTags, exif.Make, media)
	if err == nil {
		newExif.Maker = maker
	}

	lens, err := p.readStringTag(exifTags, exif.LensModel, media)
	if err == nil {
		newExif.Lens = lens
	}

	date, err := exifTags.DateTime()
	if err == nil {
		newExif.DateShot = &date
	}

	exposure, err := p.readRationalTag(exifTags, exif.ExposureTime, media)
	if err == nil {
		exposureStr := exposure.RatString()
		newExif.Exposure = &exposureStr
	}

	apertureRat, err := p.readRationalTag(exifTags, exif.FNumber, media)
	if err == nil {
		aperture, _ := apertureRat.Float64()
		newExif.Aperture = &aperture
	}

	isoTag, err := exifTags.Get(exif.ISOSpeedRatings)
	if err != nil {
		log.Printf("WARN: Could not read ISOSpeedRatings from EXIF: %s\n", media.Title)
	} else {
		iso, err := isoTag.Int(0)
		if err != nil {
			log.Printf("WARN: Could not parse EXIF ISOSpeedRatings as integer: %s\n", media.Title)
		} else {
			newExif.Iso = &iso
		}
	}

	focalLengthTag, err := exifTags.Get(exif.FocalLength)
	if err == nil {
		focalLengthRat, err := focalLengthTag.Rat(0)
		if err == nil {
			focalLength, _ := focalLengthRat.Float64()
			newExif.FocalLength = &focalLength

		} else {
			// For some photos, the focal length cannot be read as a rational value,
			// but is instead the second value read as an integer

			if err == nil {
				focalLength, err := focalLengthTag.Int(1)
				if err != nil {
					log.Printf("WARN: Could not parse EXIF FocalLength as rational or integer: %s\n%s\n", media.Title, err)
				} else {
					focalLenFloat := float64(focalLength)
					newExif.FocalLength = &focalLenFloat
				}
			}
		}
	}

	flash, err := exifTags.Flash()
	if err == nil {
		newExif.Flash = &flash
	}

	orientation, err := p.readIntegerTag(exifTags, exif.Orientation, media)
	if err == nil {
		newExif.Orientation = orientation
	}

	exposureProgram, err := p.readIntegerTag(exifTags, exif.ExposureProgram, media)
	if err == nil {
		newExif.ExposureProgram = exposureProgram
	}

	lat, long, err := exifTags.LatLong()
	if err == nil {
		newExif.GPSLatitude = &lat
		newExif.GPSLongitude = &long
	}

	returnExif = &newExif
	return
}

func (p *internalExifParser) readStringTag(tags *exif.Exif, name exif.FieldName, media *models.Media) (*string, error) {
	tag, err := tags.Get(name)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read %s from EXIF: %s", name, media.Title)
	}

	if tag != nil {
		value, err := tag.StringVal()
		if err != nil {
			return nil, errors.Wrapf(err, "could not parse %s from EXIF as string: %s", name, media.Title)
		}

		return &value, nil
	}

	log.Printf("WARN: EXIF tag %s returned null: %s\n", name, media.Title)
	return nil, errors.New("exif tag returned null")
}

func (p *internalExifParser) readRationalTag(tags *exif.Exif, name exif.FieldName, media *models.Media) (*big.Rat, error) {
	tag, err := tags.Get(name)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read %s from EXIF: %s", name, media.Title)
	}

	if tag != nil {
		value, err := tag.Rat(0)
		if err != nil {
			return nil, errors.Wrapf(err, "could not parse %s from EXIF as rational: %s", name, media.Title)
		}

		return value, nil
	}

	log.Printf("WARN: EXIF tag %s returned null: %s\n", name, media.Title)
	return nil, errors.New("exif tag returned null")
}

func (p *internalExifParser) readIntegerTag(tags *exif.Exif, name exif.FieldName, media *models.Media) (*int, error) {
	tag, err := tags.Get(name)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read %s from EXIF: %s", name, media.Title)
	}

	if tag != nil {
		value, err := tag.Int(0)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not parse %s from EXIF as integer: %s", name, media.Title)
		}

		return &value, nil
	}

	log.Printf("WARN: EXIF tag %s returned null: %s\n", name, media.Title)
	return nil, errors.New("exif tag returned null")
}
