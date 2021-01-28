package models

import (
	"time"
)

type MediaEXIF struct {
	Model
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
	GPSLongitude    *float64
}

func (MediaEXIF) TableName() string {
	return "media_exif"
}

func (exif *MediaEXIF) Media() *Media {
	panic("not implemented")
}
