package scanner

import (
	"database/sql"
	"errors"
	"log"
	"math/big"
	"os"

	"github.com/viktorstrate/photoview/api/graphql/models"
	"github.com/xor-gate/goexif2/exif"
	"github.com/xor-gate/goexif2/mknote"
)

func ScanEXIF(tx *sql.Tx, photo *models.Photo) (*models.PhotoEXIF, error) {

	log.Printf("Scanning for EXIF")

	{
		// Check if EXIF data already exists
		if photo.ExifId != nil {
			row := tx.QueryRow("SELECT * FROM photo_exif WHERE exif_id = ?", photo.ExifId)
			return models.NewPhotoExifFromRow(row)
		}

		row := tx.QueryRow("SELECT photo_exif.* FROM photo, photo_exif WHERE photo.exif_id = photo_exif.exif_id AND photo.photo_id = ?", photo.PhotoID)
		exifData, err := models.NewPhotoExifFromRow(row)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		} else if exifData != nil {
			return exifData, nil
		}
	}

	photoFile, err := os.Open(photo.Path)
	if err != nil {
		return nil, err
	}

	exif.RegisterParsers(mknote.All...)

	exifTags, err := exif.Decode(photoFile)
	if err != nil {
		return nil, err
	}

	// log.Printf("EXIF DATA FOR %s\n%s\n", photo.Title, exifTags.String())

	valueNames := make([]string, 0)
	exifValues := make([]interface{}, 0)

	model, err := readStringTag(exifTags, exif.Model, photo)
	if err == nil {
		valueNames = append(valueNames, "camera")
		exifValues = append(exifValues, model)
	}

	maker, err := readStringTag(exifTags, exif.Make, photo)
	if err == nil {
		valueNames = append(valueNames, "maker")
		exifValues = append(exifValues, maker)
	}

	lens, err := readStringTag(exifTags, exif.LensModel, photo)
	if err == nil {
		valueNames = append(valueNames, "lens")
		exifValues = append(exifValues, lens)
	}

	date, err := exifTags.DateTime()
	if err == nil {
		valueNames = append(valueNames, "dateShot")
		exifValues = append(exifValues, date)
	}

	exposure, err := readRationalTag(exifTags, exif.ExposureTime, photo)
	if err == nil {
		valueNames = append(valueNames, "exposure")
		exifValues = append(exifValues, exposure.RatString())
	}

	apertureRat, err := readRationalTag(exifTags, exif.FNumber, photo)
	if err == nil {
		aperture, _ := apertureRat.Float32()
		valueNames = append(valueNames, "aperture")
		exifValues = append(exifValues, aperture)
	}

	isoTag, err := exifTags.Get(exif.ISOSpeedRatings)
	if err != nil {
		log.Printf("WARN: Could not read ISOSpeedRatings from EXIF: %s\n", photo.Title)
	} else {
		iso, err := isoTag.Int(0)
		if err != nil {
			log.Printf("WARN: Could not parse EXIF ISOSpeedRatings as integer: %s\n", photo.Title)
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
					log.Printf("WARN: Could not parse EXIF FocalLength as rational or integer: %s\n%s\n", photo.Title, err)
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

	orientation, err := readIntegerTag(exifTags, exif.Orientation, photo)
	if err == nil {
		valueNames = append(valueNames, "orientation")
		exifValues = append(exifValues, *orientation)
	}

	exposureProgram, err := readIntegerTag(exifTags, exif.ExposureProgram, photo)
	if err == nil {
		valueNames = append(valueNames, "exposure_program")
		exifValues = append(exifValues, *exposureProgram)
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
	result, err := tx.Exec("INSERT INTO photo_exif ("+columns+") VALUES ("+prepareQuestions+")", exifValues...)
	if err != nil {
		return nil, err
	}

	exifID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Link exif to photo in database
	result, err = tx.Exec("UPDATE photo SET exif_id = ? WHERE photo_id = ?", exifID, photo.PhotoID)
	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, errors.New("Linking exif to photo in database failed: 0 rows affected")
	}

	// Return newly created exif row
	row := tx.QueryRow("SELECT * FROM photo_exif WHERE exif_id = ?", exifID)
	return models.NewPhotoExifFromRow(row)
}

func readStringTag(tags *exif.Exif, name exif.FieldName, photo *models.Photo) (*string, error) {
	tag, err := tags.Get(name)
	if err != nil {
		log.Printf("WARN: Could not read %s from EXIF: %s\n", name, photo.Title)
		return nil, err
	}

	if tag != nil {
		value, err := tag.StringVal()
		if err != nil {
			log.Printf("WARN: Could not parse %s from EXIF as string: %s\n", name, photo.Title)
			return nil, err
		}

		return &value, nil
	}

	log.Printf("WARN: EXIF tag %s returned null: %s\n", name, photo.Title)
	return nil, errors.New("exif tag returned null")
}

func readRationalTag(tags *exif.Exif, name exif.FieldName, photo *models.Photo) (*big.Rat, error) {
	tag, err := tags.Get(name)
	if err != nil {
		log.Printf("WARN: Could not read %s from EXIF: %s\n", name, photo.Title)
		return nil, err
	}

	if tag != nil {
		value, err := tag.Rat(0)
		if err != nil {
			log.Printf("WARN: Could not parse %s from EXIF as rational: %s\n%s\n", name, photo.Title, err)
			return nil, err
		}

		return value, nil
	}

	log.Printf("WARN: EXIF tag %s returned null: %s\n", name, photo.Title)
	return nil, errors.New("exif tag returned null")
}

func readIntegerTag(tags *exif.Exif, name exif.FieldName, photo *models.Photo) (*int, error) {
	tag, err := tags.Get(name)
	if err != nil {
		log.Printf("WARN: Could not read %s from EXIF: %s\n", name, photo.Title)
		return nil, err
	}

	if tag != nil {
		value, err := tag.Int(0)
		if err != nil {
			log.Printf("WARN: Could not parse %s from EXIF as integer: %s\n%s\n", name, photo.Title, err)
			return nil, err
		}

		return &value, nil
	}

	log.Printf("WARN: EXIF tag %s returned null: %s\n", name, photo.Title)
	return nil, errors.New("exif tag returned null")
}
