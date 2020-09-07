package models

import (
	"database/sql"
	"time"
)

type MediaEXIF struct {
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
	GPSLatitude     *float64
	GPSLonitude     *float64
}

func (exif *MediaEXIF) Media() *Media {
	panic("not implemented")
}

func (exif *MediaEXIF) ID() int {
	return exif.ExifID
}

func NewMediaExifFromRow(row *sql.Row) (*MediaEXIF, error) {
	exif := MediaEXIF{}

	if err := row.Scan(&exif.ExifID, &exif.Camera, &exif.Maker, &exif.Lens, &exif.DateShot, &exif.Exposure, &exif.Aperture, &exif.Iso, &exif.FocalLength, &exif.Flash, &exif.Orientation, &exif.ExposureProgram, &exif.GPSLatitude, &exif.GPSLonitude); err != nil {
		return nil, err
	}

	return &exif, nil
}
