package models

import (
	"time"
)

const (
	RFC3339Milli                = "2006-01-02T15:04:05.000Z07:00"
	RFC3339MilliWithoutTimezone = "2006-01-02T15:04:05.000"
)

type MediaEXIF struct {
	Model
	Description     *string
	Camera          *string
	Maker           *string
	Lens            *string
	DateShot        *time.Time
	DateShotStr     *string // The string of a RFC3339 time. It may or may not have a timezone, depended on if EXIF provides.
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
