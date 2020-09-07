package scanner

import (
	"database/sql"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/pkg/errors"

	"github.com/viktorstrate/photoview/api/graphql/models"
	"github.com/xor-gate/goexif2/exif"
	"github.com/xor-gate/goexif2/mknote"
)

func ScanEXIF(tx *sql.Tx, media *models.Media) (returnExif *models.MediaEXIF, returnErr error) {

	log.Printf("Scanning for EXIF")

	{
		// Check if EXIF data already exists
		if media.ExifId != nil {
			row := tx.QueryRow("SELECT * FROM media_exif WHERE exif_id = ?", media.ExifId)
			return models.NewMediaExifFromRow(row)
		}

		row := tx.QueryRow("SELECT media_exif.* FROM media, media_exif WHERE media.exif_id = media_exif.exif_id AND media.media_id = ?", media.MediaID)
		exifData, err := models.NewMediaExifFromRow(row)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		} else if exifData != nil {
			return exifData, nil
		}
	}

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

	valueNames := make([]string, 0)
	exifValues := make([]interface{}, 0)

	model, err := readStringTag(exifTags, exif.Model, media)
	if err == nil {
		valueNames = append(valueNames, "camera")
		exifValues = append(exifValues, model)
	}

	maker, err := readStringTag(exifTags, exif.Make, media)
	if err == nil {
		valueNames = append(valueNames, "maker")
		exifValues = append(exifValues, maker)
	}

	lens, err := readStringTag(exifTags, exif.LensModel, media)
	if err == nil {
		valueNames = append(valueNames, "lens")
		exifValues = append(exifValues, lens)
	}

	date, err := exifTags.DateTime()
	if err == nil {
		valueNames = append(valueNames, "date_shot")
		exifValues = append(exifValues, date)

		_, err := tx.Exec("UPDATE media SET date_shot = ? WHERE media_id = ?", date, media.MediaID)
		if err != nil {
			log.Printf("WARN: Failed to update date_shot for media %s: %s", media.Title, err)
		}
	}

	exposure, err := readRationalTag(exifTags, exif.ExposureTime, media)
	if err == nil {
		valueNames = append(valueNames, "exposure")
		exifValues = append(exifValues, exposure.RatString())
	}

	apertureRat, err := readRationalTag(exifTags, exif.FNumber, media)
	if err == nil {
		aperture, _ := apertureRat.Float32()
		valueNames = append(valueNames, "aperture")
		exifValues = append(exifValues, aperture)
	}

	isoTag, err := exifTags.Get(exif.ISOSpeedRatings)
	if err != nil {
		log.Printf("WARN: Could not read ISOSpeedRatings from EXIF: %s\n", media.Title)
	} else {
		iso, err := isoTag.Int(0)
		if err != nil {
			log.Printf("WARN: Could not parse EXIF ISOSpeedRatings as integer: %s\n", media.Title)
		} else {
			valueNames = append(valueNames, "iso")
			exifValues = append(exifValues, iso)
		}
	}

	focalLengthTag, err := exifTags.Get(exif.FocalLength)
	if err == nil {
		focalLengthRat, err := focalLengthTag.Rat(0)
		if err == nil {
			focalLength, _ := focalLengthRat.Float32()
			valueNames = append(valueNames, "focal_length")
			exifValues = append(exifValues, focalLength)
		} else {
			// For some photos, the focal length cannot be read as a rational value,
			// but is instead the second value read as an integer

			if err == nil {
				focalLength, err := focalLengthTag.Int(1)
				if err != nil {
					log.Printf("WARN: Could not parse EXIF FocalLength as rational or integer: %s\n%s\n", media.Title, err)
				} else {
					valueNames = append(valueNames, "focal_length")
					exifValues = append(exifValues, focalLength)
				}
			}
		}
	}

	flash, err := exifTags.Flash()
	if err == nil {
		valueNames = append(valueNames, "flash")
		exifValues = append(exifValues, flash)
	}

	orientation, err := readIntegerTag(exifTags, exif.Orientation, media)
	if err == nil {
		valueNames = append(valueNames, "orientation")
		exifValues = append(exifValues, *orientation)
	}

	exposureProgram, err := readIntegerTag(exifTags, exif.ExposureProgram, media)
	if err == nil {
		valueNames = append(valueNames, "exposure_program")
		exifValues = append(exifValues, *exposureProgram)
	}

	lat, long, err := exifTags.LatLong()
	if err == nil {
		valueNames = append(valueNames, "gps_latitude")
		exifValues = append(exifValues, lat)

		valueNames = append(valueNames, "gps_longitude")
		exifValues = append(exifValues, long)
	}

	if len(valueNames) == 0 {
		return nil, nil
	}

	prepareQuestions := ""
	for range valueNames {
		prepareQuestions += "?,"
	}
	prepareQuestions = prepareQuestions[0 : len(prepareQuestions)-1]

	columns := ""
	for _, name := range valueNames {
		columns += name + ","
	}
	columns = columns[0 : len(columns)-1]

	// Insert into database
	result, err := tx.Exec("INSERT INTO media_exif ("+columns+") VALUES ("+prepareQuestions+")", exifValues...)
	if err != nil {
		return nil, err
	}

	exifID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Link exif to media in database
	result, err = tx.Exec("UPDATE media SET exif_id = ? WHERE media_id = ?", exifID, media.MediaID)
	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, errors.Wrap(err, "linking exif to media in database failed")
	}

	if rowsAffected == 0 {
		return nil, errors.New("linking exif to media in database failed: 0 rows affected")
	}

	// Return newly created exif row
	row := tx.QueryRow("SELECT * FROM media_exif WHERE exif_id = ?", exifID)
	return models.NewMediaExifFromRow(row)
}

func readStringTag(tags *exif.Exif, name exif.FieldName, media *models.Media) (*string, error) {
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

func readRationalTag(tags *exif.Exif, name exif.FieldName, media *models.Media) (*big.Rat, error) {
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

func readIntegerTag(tags *exif.Exif, name exif.FieldName, media *models.Media) (*int, error) {
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
