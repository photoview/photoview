package models

import (
	"time"
)

type MediaEXIF struct {
	Model
	Description     *string
	Camera          *string
	Maker           *string
	Lens            *string
	DateShot        *time.Time
	Exposure        *float64
	Aperture        *float64
	Iso             *int64
	FocalLength     *float64
	Flash           *int64
	Orientation     *int64
	ExposureProgram *int64
	GPSLatitude     *float64
	GPSLongitude    *float64
}

func (MediaEXIF) TableName() string {
	return "media_exif"
}

func (exif *MediaEXIF) Media() *Media {
	panic("not implemented")
}

func (exif *MediaEXIF) Coordinates() *Coordinates {
	if exif.GPSLatitude == nil || exif.GPSLongitude == nil {
		return nil
	}

	return &Coordinates{
		Latitude:  *exif.GPSLatitude,
		Longitude: *exif.GPSLongitude,
	}
}
