package database

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type exifModel struct {
	ID       int `gorm:"primarykey"`
	Exposure *string
	Flash    *string
}

var flashDescriptions = map[int]string{
	0x0:  "No Flash",
	0x1:  "Fired",
	0x5:  "Fired, Return not detected",
	0x7:  "Fired, Return detected",
	0x8:  "On, Did not fire",
	0x9:  "On, Fired",
	0xD:  "On, Return not detected",
	0xF:  "On, Return detected",
	0x10: "Off, Did not fire",
	0x14: "Off, Did not fire, Return not detected",
	0x18: "Auto, Did not fire",
	0x19: "Auto, Fired",
	0x1D: "Auto, Fired, Return not detected",
	0x1F: "Auto, Fired, Return detected",
	0x20: "No flash function",
	0x30: "Off, No flash function",
	0x41: "Fired, Red-eye reduction",
	0x45: "Fired, Red-eye reduction, Return not detected",
	0x47: "Fired, Red-eye reduction, Return detected",
	0x49: "On, Red-eye reduction",
	0x4D: "On, Red-eye reduction, Return not detected",
	0x4F: "On, Red-eye reduction, Return detected",
	0x50: "Off, Red-eye reduction",
	0x58: "Auto, Did not fire, Red-eye reduction",
	0x59: "Auto, Fired, Red-eye reduction",
	0x5D: "Auto, Fired, Red-eye reduction, Return not detected",
	0x5F: "Auto, Fired, Red-eye reduction, Return detected",
}

// Migrate MediaExif fields "exposure" and "flash" from strings to integers
func migrateExifFields(db *gorm.DB) error {
	mediaExifColumns, err := db.Migrator().ColumnTypes(&models.MediaEXIF{})
	if err != nil {
		return err
	}

	return db.Transaction(func(tx *gorm.DB) error {
		for _, exifCol := range mediaExifColumns {
			if err := processExposure(exifCol, db); err != nil {
				return err
			}

			if err := processFlash(exifCol, db); err != nil {
				return err
			}
		}

		if err := db.AutoMigrate(&models.MediaEXIF{}); err != nil {
			return errors.Wrap(err, "failed to auto migrate media_exif after exposure conversion")
		}

		return nil
	})
}

func processFlash(exifCol gorm.ColumnType, db *gorm.DB) error {
	if exifCol.Name() == "flash" {
		switch exifCol.DatabaseTypeName() {
		case "double", "numeric", "real", "bigint", "integer":
			// correct type, do nothing
		default:
			// do migration
			if err := migrateExifFieldsFlash(db); err != nil {
				return err
			}
		}
	}
	return nil
}

func processExposure(exifCol gorm.ColumnType, db *gorm.DB) error {
	if exifCol.Name() == "exposure" {
		switch exifCol.DatabaseTypeName() {
		case "double", "numeric", "real", "bigint", "integer":
			// correct type, do nothing
		default:
			// do migration
			if err := migrateExifFieldsExposure(db); err != nil {
				return err
			}
		}
	}
	return nil
}

func migrateExifFieldsExposure(db *gorm.DB) error {
	log.Println("Migrating `media_exif.exposure` from string to double")

	err := db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Exec("UPDATE media_exif SET exposure = NULL WHERE exposure = ''").Error; err != nil {
			return errors.Wrapf(err, "convert flash attribute empty values to NULL")
		}

		var results []exifModel

		return calculateExposure(tx, results)
	})

	if err != nil {
		return errors.Wrap(err, "migrating `media_exif.exposure` failed")
	}

	return nil
}

func calculateExposure(tx *gorm.DB, results []exifModel) error {
	return tx.Model(&exifModel{}).Table("media_exif").Where("exposure LIKE '%/%'").FindInBatches(
		&results, 100, func(tx *gorm.DB, batch int) error {
			for _, result := range results {

				if result.Exposure == nil {
					continue
				}

				frac := strings.Split(*result.Exposure, "/")
				if len(frac) != 2 {
					return errors.Errorf("failed to convert exposure value (%s) expected format x/y", frac)
				}

				numerator, err := strconv.ParseFloat(frac[0], 64)
				if err != nil {
					return err
				}

				denominator, err := strconv.ParseFloat(frac[1], 64)
				if err != nil {
					return err
				}

				decimalValue := numerator / denominator
				*result.Exposure = fmt.Sprintf("%f", decimalValue)
			}

			tx.Save(&results)

			return nil
		}).Error
}

func migrateExifFieldsFlash(db *gorm.DB) error {
	log.Println("Migrating `media_exif.flash` from string to int")

	err := db.Transaction(func(tx *gorm.DB) error {

		var dataType string
		if err := tx.Raw(
			"SELECT data_type FROM information_schema.columns WHERE table_name = 'media_exif' AND column_name = 'flash';").
			Find(&dataType).Error; err != nil {

			return errors.Wrapf(err, "read data_type of column media_exif.flash")
		}

		if dataType == "bigint" {
			return nil
		}

		if err := tx.Exec("UPDATE media_exif SET flash = NULL WHERE flash = ''").Error; err != nil {
			return errors.Wrapf(err, "convert flash attribute empty values to NULL")
		}

		var results []exifModel

		return replaceFlashValues(tx, results)
	})

	if err != nil {
		return errors.Wrap(err, "migrating `media_exif.flash` failed")
	}

	return nil
}

func replaceFlashValues(tx *gorm.DB, results []exifModel) error {
	return tx.Model(&exifModel{}).Table("media_exif").Where("flash IS NOT NULL").FindInBatches(
		&results, 100, func(tx *gorm.DB, batch int) error {
			for _, result := range results {

				if result.Flash == nil {
					continue
				}

				for index, name := range flashDescriptions {
					if *result.Flash == name {
						*result.Flash = fmt.Sprintf("%d", index)
						break
					}
				}
			}

			tx.Save(&results)

			return nil
		}).Error
}
