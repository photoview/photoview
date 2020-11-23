package models

import (
	"time"

	"gorm.io/gorm"
)

type MediaEXIF struct {
	gorm.Model
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
