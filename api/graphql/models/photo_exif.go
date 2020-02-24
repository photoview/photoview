package models

import (
	"database/sql"
	"time"
)

type PhotoEXIF struct {
	ExifID          int
	Camera          *string
	Maker           *string
	Lens            *string
	DateShot        *time.Time
	Exposure        *string
	Aperture        *float64
	Iso             *int
	FocalLength     *float64
	Flash           *string
	Orientation     *int
	ExposureProgram *int
}

func (exif *PhotoEXIF) Photo() *Photo {
	panic("not implemented")
}

func (exif *PhotoEXIF) ID() int {
	return exif.ExifID
}

func NewPhotoExifFromRow(row *sql.Row) (*PhotoEXIF, error) {
	exif := PhotoEXIF{}

	if err := row.Scan(&exif.ExifID, &exif.Camera, &exif.Maker, &exif.Lens, &exif.DateShot, &exif.Exposure, &exif.Aperture, &exif.Iso, &exif.FocalLength, &exif.Flash, &exif.Orientation, &exif.ExposureProgram); err != nil {
		return nil, err
	}

	return &exif, nil
}
