package exif

import (
	"fmt"
	"log"
	"math"
	"math/big"
	"os"
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/pkg/errors"
	"github.com/xor-gate/goexif2/exif"
	"github.com/xor-gate/goexif2/mknote"
)

// internalExifParser is an exif parser that parses the media without the use of external tools
type internalExifParser struct{}

const couldNotReadXfromEXIFy = "could not read %s from EXIF: %s"
const warnEXIFtagXreturnedNully = "WARN: EXIF tag %s returned null: %s\n"

var ErrNullExifTag = errors.New("exif tag returned null")

func NewInternalExifParser() ExifParser {
	return internalExifParser{}
}

func (p internalExifParser) ParseExif(mediaPath string) (returnExif *models.MediaEXIF, returnErr error) {
	photoFile, err := os.Open(mediaPath)
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
		return nil, nil
		// return nil, errors.Wrap(err, "Could not decode EXIF")
	}

	newExif := models.MediaEXIF{}

	description, err := p.readStringTag(exifTags, exif.ImageDescription, mediaPath)
	if err == nil {
		newExif.Description = description
	}

	model, err := p.readStringTag(exifTags, exif.Model, mediaPath)
	if err == nil {
		newExif.Camera = model
	}

	maker, err := p.readStringTag(exifTags, exif.Make, mediaPath)
	if err == nil {
		newExif.Maker = maker
	}

	lens, err := p.readStringTag(exifTags, exif.LensModel, mediaPath)
	if err == nil {
		newExif.Lens = lens
	}

	date, err := exifTags.DateTime()
	if err == nil {
		_, tz := date.Zone()
		dateUTC := date.Add(time.Duration(tz) * time.Second).UTC()
		newExif.DateShot = &dateUTC
	}

	exposureTag, err := exifTags.Get(exif.ExposureTime)
	if err == nil {
		exposureRat, err := exposureTag.Rat(0)
		if err == nil {
			exposure, _ := exposureRat.Float64()
			newExif.Exposure = &exposure
		}
	}

	apertureRat, err := p.readRationalTag(exifTags, exif.FNumber, mediaPath)
	if err == nil {
		aperture, _ := apertureRat.Float64()
		newExif.Aperture = &aperture
	}

	isoTag, err := exifTags.Get(exif.ISOSpeedRatings)
	if err != nil {
		log.Printf("WARN: Could not read ISOSpeedRatings from EXIF: %v\n", mediaPath)
	} else {
		iso, err := isoTag.Int(0)
		if err != nil {
			log.Printf("WARN: Could not parse EXIF ISOSpeedRatings as integer: %v\n", mediaPath)
		} else {
			iso64 := int64(iso)
			newExif.Iso = &iso64
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
			focalLength, err := focalLengthTag.Int(1)
			if err != nil {
				log.Printf("WARN: Could not parse EXIF FocalLength as rational or integer: %v\n%s\n", mediaPath, err)
			} else {
				focalLenFloat := float64(focalLength)
				newExif.FocalLength = &focalLenFloat
			}
		}
	}

	flash, err := p.readIntegerTag(exifTags, exif.Flash, mediaPath)
	if err == nil {
		flash64 := int64(*flash)
		newExif.Flash = &flash64
	}

	orientation, err := p.readIntegerTag(exifTags, exif.Orientation, mediaPath)
	if err == nil {
		orientation64 := int64(*orientation)
		newExif.Orientation = &orientation64
	}

	exposureProgram, err := p.readIntegerTag(exifTags, exif.ExposureProgram, mediaPath)
	if err == nil {
		exposureProgram64 := int64(*exposureProgram)
		newExif.ExposureProgram = &exposureProgram64
	}

	lat, long, err := exifTags.LatLong()
	if err == nil {
		if math.Abs(lat) > 90 || math.Abs(long) > 90 {
			returnExif = &newExif
			log.Printf(
				"Incorrect GPS data in the %s Exif data: %f, %f, while expected values between '-90' and '90'. Ignoring GPS data.",
				mediaPath, long, lat)
			return
		} else {
			newExif.GPSLatitude = &lat
			newExif.GPSLongitude = &long
		}
	}

	returnExif = &newExif
	return
}

func (p *internalExifParser) readStringTag(tags *exif.Exif, name exif.FieldName, mediaPath string) (*string, error) {
	tag, err := tags.Get(name)
	if err != nil {
		return nil, errors.Wrapf(err, couldNotReadXfromEXIFy, name, mediaPath)
	}

	if tag != nil {
		value, err := tag.StringVal()
		if err != nil {
			return nil, errors.Wrapf(err, "could not parse %s from EXIF as string: %s", name, mediaPath)
		}

		return &value, nil
	}

	log.Printf(warnEXIFtagXreturnedNully, name, mediaPath)
	return nil, ErrNullExifTag
}

func (p *internalExifParser) readRationalTag(tags *exif.Exif, name exif.FieldName, mediaPath string) (*big.Rat, error) {
	tag, err := tags.Get(name)
	if err != nil {
		return nil, errors.Wrapf(err, couldNotReadXfromEXIFy, name, mediaPath)
	}

	if tag != nil {
		value, err := tag.Rat(0)
		if err != nil {
			return nil, errors.Wrapf(err, "could not parse %s from EXIF as rational: %s", name, mediaPath)
		}

		return value, nil
	}

	log.Printf(warnEXIFtagXreturnedNully, name, mediaPath)
	return nil, ErrNullExifTag
}

func (p *internalExifParser) readIntegerTag(tags *exif.Exif, name exif.FieldName, mediaPath string) (*int, error) {
	tag, err := tags.Get(name)
	if err != nil {
		return nil, errors.Wrapf(err, couldNotReadXfromEXIFy, name, mediaPath)
	}

	if tag != nil {
		value, err := tag.Int(0)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not parse %s from EXIF as integer: %s", name, mediaPath)
		}

		return &value, nil
	}

	log.Printf(warnEXIFtagXreturnedNully, name, mediaPath)
	return nil, ErrNullExifTag
}
